package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type KVExport struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	backupFlag := flag.Bool("backup", false, "Backup local Consul KV")
	exportFlag := flag.Bool("export", false, "Export staging KV to consul_export.json")
	importFlag := flag.Bool("import", false, "Backup local + import staging into local")

	fromPrefixFlag := flag.String("from-prefix", "", "Only import keys under this prefix (default root)")
	toPrefixFlag := flag.String("to-prefix", "", "Rewrite imported keys to this prefix (default root)")

	flag.Parse()

	if !*backupFlag && !*exportFlag && !*importFlag {
		fmt.Println("Usage:")
		fmt.Println("  consul-sync --backup")
		fmt.Println("  consul-sync --export")
		fmt.Println("  consul-sync --import")
		fmt.Println("")
		fmt.Println("Import options:")
		fmt.Println("  --from-prefix=\"serviceA/\"")
		fmt.Println("  --to-prefix=\"localA/\"")
		os.Exit(1)
	}

	if *backupFlag {
		backupLocal()
	}

	if *exportFlag {
		exportStaging()
	}

	if *importFlag {
		backupLocal() // safe before changes
		importToLocal(*fromPrefixFlag, *toPrefixFlag)
	}
}

//
// --------------------------------------------------------------
//  BACKUP LOCAL CONSUL
// --------------------------------------------------------------
//

func backupLocal() {
	localAddr := getenv("LOCAL_CONSUL_ADDR", "http://localhost:8500")
	localToken := os.Getenv("LOCAL_CONSUL_TOKEN")

	client, err := newClient(localAddr, localToken)
	if err != nil {
		log.Fatalf("Failed to create local client: %v", err)
	}

	keys, _, err := client.KV().Keys("", "", nil)
	if err != nil {
		log.Fatalf("Failed to list local keys: %v", err)
	}

	export := []KVExport{}
	for _, key := range keys {
		kv, _, err := client.KV().Get(key, nil)
		if err != nil {
			log.Printf("⚠️ Failed to read key %s: %v", key, err)
			continue
		}
		if kv != nil {
			export = append(export, KVExport{
				Key:   key,
				Value: string(kv.Value),
			})
		}
	}

	filename := "local_backup_" + time.Now().Format("2006-01-02_150405") + ".json"
	data, _ := json.MarshalIndent(export, "", "  ")
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write backup: %v", err)
	}

	fmt.Printf("✅ Local backup created: %s (%d keys)\n", filename, len(export))
}

//
// --------------------------------------------------------------
//  EXPORT STAGING → consul_export.json
// --------------------------------------------------------------
//

func exportStaging() {
	stagingAddr := getenv("STAGING_CONSUL_ADDR", "http://staging-consul:8500")
	stagingToken := os.Getenv("STAGING_CONSUL_TOKEN")

	client, err := newClient(stagingAddr, stagingToken)
	if err != nil {
		log.Fatalf("Failed to create staging client: %v", err)
	}

	keys, _, err := client.KV().Keys("", "", nil)
	if err != nil {
		log.Fatalf("Failed to list staging keys: %v", err)
	}

	export := []KVExport{}
	for _, key := range keys {
		kv, _, err := client.KV().Get(key, nil)
		if err != nil {
			log.Printf("⚠️ Failed to read key %s: %v", key, err)
			continue
		}
		if kv != nil {
			export = append(export, KVExport{
				Key:   key,
				Value: string(kv.Value),
			})
		}
	}

	data, _ := json.MarshalIndent(export, "", "  ")
	err = os.WriteFile("consul_export.json", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write export file: %v", err)
	}

	fmt.Printf("✅ Exported staging KV → consul_export.json (%d keys)\n", len(export))
}

//
// --------------------------------------------------------------
//  IMPORT (PREFIX FILTER + SAFE PREFIX REWRITE)
// --------------------------------------------------------------
//

func importToLocal(fromPrefix string, toPrefix string) {
	localAddr := getenv("LOCAL_CONSUL_ADDR", "http://localhost:8500")
	localToken := os.Getenv("LOCAL_CONSUL_TOKEN")

	client, err := newClient(localAddr, localToken)
	if err != nil {
		log.Fatalf("Failed to create local client: %v", err)
	}

	file, err := os.ReadFile("consul_export.json")
	if err != nil {
		log.Fatalf("Failed to read consul_export.json: %v", err)
	}

	var data []KVExport
	if err := json.Unmarshal(file, &data); err != nil {
		log.Fatalf("Failed to parse consul_export.json: %v", err)
	}

	// Filter keys by from-prefix
	filtered := []KVExport{}
	for _, item := range data {
		if fromPrefix == "" || strings.HasPrefix(item.Key, fromPrefix) {
			filtered = append(filtered, item)
		}
	}

	fmt.Printf("📥 Importing %d keys (from-prefix='%s' → to-prefix='%s')...\n",
		len(filtered), fromPrefix, toPrefix)

	for i, item := range filtered {
		originalKey := item.Key

		//---------------------------------
		// 1. Strip from-prefix safely
		//---------------------------------
		rewrittenKey := originalKey
		if fromPrefix != "" && strings.HasPrefix(originalKey, fromPrefix) {
			rewrittenKey = originalKey[len(fromPrefix):]
		}

		//---------------------------------
		// 2. Apply to-prefix safely
		//---------------------------------
		if toPrefix != "" {
			// ensure we don't create "/xxx"
			trimmedTo := strings.TrimLeft(toPrefix, "/")
			rewrittenKey = trimmedTo + rewrittenKey
		}

		//---------------------------------
		// 3. Skip empty keys
		//---------------------------------
		if rewrittenKey == "" {
			log.Printf("⚠️ Skipping key '%s' because it becomes an empty key", originalKey)
			continue
		}

		//---------------------------------
		// 4. Remove any leading slash
		//---------------------------------
		rewrittenKey = strings.TrimLeft(rewrittenKey, "/")

		_, err = client.KV().Put(&api.KVPair{
			Key:   rewrittenKey,
			Value: []byte(item.Value),
		}, nil)

		if err != nil {
			log.Printf("[%d/%d] ❌ Failed: %s → %s : %v",
				i+1, len(filtered), originalKey, rewrittenKey, err)
		} else {
			log.Printf("[%d/%d] ✅ %s → %s",
				i+1, len(filtered), originalKey, rewrittenKey)
		}
	}

	fmt.Println("🎉 Import complete!")
}

//
// --------------------------------------------------------------
//  HELPERS
// --------------------------------------------------------------
//

func newClient(addr, token string) (*api.Client, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr
	if token != "" {
		cfg.Token = token
	}
	return api.NewClient(cfg)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
