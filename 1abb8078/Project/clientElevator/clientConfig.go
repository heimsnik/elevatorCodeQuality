package clientElevator

import "time"

const _pollAliveRate = 100 * time.Millisecond

const _pollLightRate = 20 * time.Millisecond

const _pollConnectionCheckRate = 20 * time.Millisecond

const _primarySpawnTime = 5 * time.Second

const _reattemptConnectToPrimaryTime = 3 * time.Second