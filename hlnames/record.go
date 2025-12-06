package hlnames

import (
	"encoding/json"
	"log"
	"net/http"
)

func QueryKV(nameHash string) map[string]interface{} {
	// query from https://api.hlnames.xyz/records/data_record/{nameHash}
	// with header X-API-Key: CPEPKMI-HUSUX6I-SE2DHEA-YYWFG5Y

	req, err := http.NewRequest("GET", "https://api.hlnames.xyz/records/data_record/"+nameHash, nil)
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil
	}
	req.Header.Set("X-API-Key", "CPEPKMI-HUSUX6I-SE2DHEA-YYWFG5Y")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected HTTP status: %d", resp.StatusCode)
		return nil
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("Error decoding JSON response: %v", err)
		return nil
	}

	records, ok := result["records"].(map[string]interface{})
	if !ok {
		log.Printf("Error: 'records' field not found or invalid type")
		return nil
	}

	return records
}
