package rtdb

import (
	"log"
	"os"
)

var (
	rtdbLogger = log.New(os.Stdout, "[rtdb] ", log.Ldate|log.Lshortfile|log.Ltime|log.Ldate)
)
