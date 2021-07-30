package helpers

type PlanisphereReportLookupOp struct {
	PlatormLookups *PlatformTool
}

type PlatformTool interface {
	Memory() (uint64, error)
}
