package primary

import "time"

const _pollRateSendAlive = 20 * time.Millisecond
const _pollRateBackup = 1 * time.Second
const _pollRateTimeout = 100 * time.Millisecond
const _maxTimeOnRequest = 10 * time.Second
const _maxSendRate = 100 * time.Millisecond
const _minTimeBetweenSpawnBackup = 8 * time.Second
const _pollPrimaryAliveRate = 5 * time.Second
const _pollPrimaryAliveSubtick = 100 * time.Millisecond
const _pollPrimaryAliveMaxFails = 10