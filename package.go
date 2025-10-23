// Package snowflake is a network service for generating unique ID numbers at high scale with some simple guarantees.
// The first bit is unused sign bit.
// The second part consists of a 48-bit timestamp (milliseconds) whose value is the offset of the current time relative to a certain time.
// The 7 bits machineID, max value is 2^7 -1 = 127.
// The last part consists of 9 bits, its means the length of the serial number generated per millisecond per working node, a maximum of 2^9 -1 = 511 IDs can be generated in the same millisecond.
// In a distributed environment, seven-bit machineID means that can deploy up to 127 machines.
// The binary length of 48 bits is at most 2^48 -1 millisecond = 8,925 years. So the snowflake algorithm can be used for up to 8,925 years, In order to maximize the use of the algorithm, you should specify a start time for it.
package snowflake
