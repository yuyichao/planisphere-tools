/*
Package hardware defines some information about hardware data. This should probably be moved elsewhere, it doesn't need it's own module
*/
package hardware

// ChassisType maps integers to planisphere server types
var ChassisType = map[int]string{
	3:  "desktop",
	4:  "desktop",
	5:  "desktop",
	6:  "desktop",
	7:  "desktop",
	8:  "laptop",
	9:  "laptop",
	10: "laptop",
	11: "laptop",
	13: "desktop",
	14: "laptop",
	17: "server_physical",
	21: "laptop",
	23: "server_physical",
	25: "server_physical",
}

// RaspberryPiModels maps the bios number to actual model numbers
var RaspberryPiModels = map[string]string{
	"0002":   "B",
	"0003":   "B",
	"0004":   "B",
	"0005":   "B",
	"0006":   "B",
	"0007":   "A",
	"0008":   "A",
	"0009":   "A",
	"000d":   "B",
	"000e":   "B",
	"000f":   "B",
	"0010":   "B+",
	"0011":   "Compute Module 1",
	"0012":   "A+",
	"0013":   "B+",
	"0014":   "Compute Module 1",
	"0015":   "A+",
	"a01040": "2 Model B",
	"a21041": "2 Model B",
	"a22042": "2 Model B",
	"900021": "A+",
	"900032": "B+",
	"900092": "Zero",
	"900093": "Zero",
	"920093": "Zero",
	"9000c1": "Zero W",
	"a02082": "3 Model B",
	"a020a0": "Compute Module 3",
	"a22082": "3 Model B",
	"a32082": "3 Model B",
}
