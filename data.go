package main

import "time"

type DvLog struct {
	VideoData time.Time `gorm:"VIDEO_DATE"`
	Comment   string    `gorm:"LOG_COMMENT"`
}
