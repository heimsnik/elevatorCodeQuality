package backup

import "time"

const _pollPrimaryAlive = 100 * time.Millisecond
const _pollSendAlive = 20 * time.Millisecond
const _connectTimeout = 6 * time.Second
