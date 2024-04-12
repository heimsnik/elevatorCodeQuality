package elevio

import "time"

const NUM_FLOORS int = 4
const NUM_BUTTONS int = 3
const _pollRate = 20 * time.Millisecond
var _initialized bool = false