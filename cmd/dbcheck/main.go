package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".config", "netpulse", "netpulse.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	fmt.Println("=== PROBE SUMMARY (last 30 min) ===")
	rows, _ := db.Query(`
		SELECT probe_type, COUNT(*) as cnt,
			ROUND(AVG(latency_ms),1) as avg_ms,
			ROUND(MIN(latency_ms),1) as min_ms,
			ROUND(MAX(latency_ms),1) as max_ms,
			ROUND(AVG(CASE WHEN success=1 THEN 0.0 ELSE 1.0 END)*100,1) as fail_pct
		FROM probe_results
		WHERE timestamp > datetime('now','-30 minutes')
		GROUP BY probe_type
		ORDER BY probe_type`)
	defer rows.Close()
	fmt.Printf("%-10s %6s %8s %8s %8s %8s\n", "TYPE", "COUNT", "AVG(ms)", "MIN(ms)", "MAX(ms)", "FAIL%")
	for rows.Next() {
		var pt string
		var cnt int
		var avg, min, max, fail float64
		rows.Scan(&pt, &cnt, &avg, &min, &max, &fail)
		fmt.Printf("%-10s %6d %8.1f %8.1f %8.1f %8.1f\n", pt, cnt, avg, min, max, fail)
	}

	fmt.Println("\n=== LATENCY BY TARGET (last 5 min) ===")
	rows2, _ := db.Query(`
		SELECT probe_type, target,
			ROUND(AVG(latency_ms),1),
			ROUND(AVG(CASE WHEN success=1 THEN 0.0 ELSE 1.0 END)*100,1)
		FROM probe_results
		WHERE timestamp > datetime('now','-5 minutes') AND probe_type != 'wifi'
		GROUP BY probe_type, target
		ORDER BY AVG(latency_ms) DESC`)
	defer rows2.Close()
	fmt.Printf("%-10s %-30s %8s %8s\n", "TYPE", "TARGET", "AVG(ms)", "FAIL%")
	for rows2.Next() {
		var pt, target string
		var avg, fail float64
		rows2.Scan(&pt, &target, &avg, &fail)
		fmt.Printf("%-10s %-30s %8.1f %8.1f\n", pt, target, avg, fail)
	}

	fmt.Println("\n=== WIFI (last 5 snapshots) ===")
	rows3, _ := db.Query(`SELECT timestamp, interface, ssid, signal_dbm, noise_dbm, channel, link_speed_mbps FROM wifi_snapshots ORDER BY timestamp DESC LIMIT 5`)
	defer rows3.Close()
	for rows3.Next() {
		var ts, iface, ssid string
		var signal, noise, ch int
		var speed float64
		rows3.Scan(&ts, &iface, &ssid, &signal, &noise, &ch, &speed)
		fmt.Printf("  %s | %s | SSID: %s | Signal: %ddBm | Noise: %ddBm | Ch: %d | Speed: %.0f Mbps\n", ts, iface, ssid, signal, noise, ch, speed)
	}

	fmt.Println("\n=== SPEED TESTS ===")
	rows4, _ := db.Query(`SELECT timestamp, download_mbps, upload_mbps, latency_ms, jitter_ms FROM speed_tests ORDER BY timestamp DESC LIMIT 5`)
	defer rows4.Close()
	for rows4.Next() {
		var ts string
		var dl, ul, lat, jit float64
		rows4.Scan(&ts, &dl, &ul, &lat, &jit)
		fmt.Printf("  %s | ↓%.1f Mbps | ↑%.1f Mbps | Latency: %.0fms | Jitter: %.0fms\n", ts, dl, ul, lat, jit)
	}

	fmt.Println("\n=== TOTALS ===")
	var total int
	db.QueryRow(`SELECT COUNT(*) FROM probe_results`).Scan(&total)
	fmt.Printf("  Probe results: %d\n", total)
	db.QueryRow(`SELECT COUNT(*) FROM diagnoses`).Scan(&total)
	fmt.Printf("  Diagnoses: %d\n", total)
	db.QueryRow(`SELECT COUNT(*) FROM speed_tests`).Scan(&total)
	fmt.Printf("  Speed tests: %d\n", total)
	db.QueryRow(`SELECT COUNT(*) FROM wifi_snapshots`).Scan(&total)
	fmt.Printf("  Wifi snapshots: %d\n", total)
}
