// Package snowflake is a network service for generating unique ID numbers at high scale with some simple guarantees.
// The first bit is unused sign bit.
// The second part consists of a 43-bit timestamp (milliseconds) whose value is the offset of the current time relative to a certain time.
// The 9 bits machineID, max value is 2^9 -1 = 511.
// The last part consists of 12 bits, its means the length of the serial number generated per millisecond per working node, a maximum of 2^12 -1 = 4095 IDs can be generated in the same millisecond.
// In a distributed environment, nine-bit machineID means that can deploy up to 511 machines.
// The binary length of 43 bits is at most 2^43 -1 millisecond = 279 years. So the snowflake algorithm can be used for up to 279 years, In order to maximize the use of the algorithm, you should specify a start time for it.
package snowflake
