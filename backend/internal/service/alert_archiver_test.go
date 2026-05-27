package service

import (
	"testing"
)

// TestDefaultAlertArchiverConfig tests DefaultAlertArchiverConfig function
func TestDefaultAlertArchiverConfig(t *testing.T) {
	cfg := DefaultAlertArchiverConfig()

	if cfg.ArchiveDaysOld <= 0 {
		t.Error("ArchiveDaysOld should be positive")
	}
	if cfg.DeleteDaysOld <= 0 {
		t.Error("DeleteDaysOld should be positive")
	}
	if cfg.DeleteDaysOld < cfg.ArchiveDaysOld {
		t.Error("DeleteDaysOld should >= ArchiveDaysOld")
	}
	if cfg.ScheduleIntervalHours <= 0 {
		t.Error("ScheduleIntervalHours should be positive")
	}
}
