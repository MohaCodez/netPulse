package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/amit/netpulse/internal/storage"
)

// Format specifies the export format.
type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
)

// Options configures what to export.
type Options struct {
	Format    Format
	Since     time.Time
	Until     time.Time
	ProbeType string // filter by probe type, empty = all
}

// ExportProbeResults writes probe results to the given writer.
func ExportProbeResults(w io.Writer, results []storage.ProbeResult, format Format) error {
	switch format {
	case FormatJSON:
		return exportProbeResultsJSON(w, results)
	case FormatCSV:
		return exportProbeResultsCSV(w, results)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ExportDiagnoses writes diagnosis history to the given writer.
func ExportDiagnoses(w io.Writer, diagnoses []storage.Diagnosis, format Format) error {
	switch format {
	case FormatJSON:
		return exportDiagnosesJSON(w, diagnoses)
	case FormatCSV:
		return exportDiagnosesCSV(w, diagnoses)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ExportSpeedTests writes speed test results to the given writer.
func ExportSpeedTests(w io.Writer, tests []storage.SpeedTestResult, format Format) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(tests)
	case FormatCSV:
		return exportSpeedTestsCSV(w, tests)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func exportProbeResultsJSON(w io.Writer, results []storage.ProbeResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func exportProbeResultsCSV(w io.Writer, results []storage.ProbeResult) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	if err := writer.Write([]string{
		"timestamp", "probe_type", "target", "success", "latency_ms", "jitter_ms", "packet_loss",
	}); err != nil {
		return err
	}

	for _, r := range results {
		success := "false"
		if r.Success {
			success = "true"
		}
		if err := writer.Write([]string{
			r.Timestamp.Format(time.RFC3339),
			r.ProbeType,
			r.Target,
			success,
			fmt.Sprintf("%.2f", r.LatencyMs),
			fmt.Sprintf("%.2f", r.JitterMs),
			fmt.Sprintf("%.4f", r.PacketLoss),
		}); err != nil {
			return err
		}
	}

	return nil
}

func exportDiagnosesJSON(w io.Writer, diagnoses []storage.Diagnosis) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(diagnoses)
}

func exportDiagnosesCSV(w io.Writer, diagnoses []storage.Diagnosis) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{
		"timestamp", "category", "severity", "title", "description", "resolved", "resolved_at",
	}); err != nil {
		return err
	}

	for _, d := range diagnoses {
		resolved := "false"
		if d.Resolved {
			resolved = "true"
		}
		resolvedAt := ""
		if d.ResolvedAt != nil {
			resolvedAt = d.ResolvedAt.Format(time.RFC3339)
		}
		if err := writer.Write([]string{
			d.Timestamp.Format(time.RFC3339),
			d.Category,
			d.Severity,
			d.Title,
			d.Description,
			resolved,
			resolvedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

func exportSpeedTestsCSV(w io.Writer, tests []storage.SpeedTestResult) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{
		"timestamp", "download_mbps", "upload_mbps", "latency_ms", "jitter_ms", "server", "triggered_by",
	}); err != nil {
		return err
	}

	for _, t := range tests {
		if err := writer.Write([]string{
			t.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", t.DownloadMbps),
			fmt.Sprintf("%.2f", t.UploadMbps),
			fmt.Sprintf("%.2f", t.LatencyMs),
			fmt.Sprintf("%.2f", t.JitterMs),
			t.Server,
			t.TriggeredBy,
		}); err != nil {
			return err
		}
	}

	return nil
}
