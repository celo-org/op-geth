package params

// Cel2Config holds config required when running a cel2 chain.
type Cel2Config struct {
	// TransitionBlockBaseFee is the base fee for the transition block in a
	// migrated chain. it must be set for migrated chains during the migration
	// to the value in the transition block. For non migrated chains it does
	// not need to be set.This is required because we need
	// eip1559.CalcBaseFee(config *params.ChainConfig, parent *types.Header,
	// time uint64) to be able to return the correct base fee for the
	// transition block and CalcBaseFee does not have access to the current
	// header so cannot know what the base fee should be. We can't just use the
	// base fee of the parent either because if we are transitioning at a
	// pre-gingerbread block then it won't have a base fee, so this seems like
	// the least invasive approach. Alternatively we could change the signature
	// of CalcBaseFee to include the current header but that would require
	// changing code in a number of places.
	TransitionBlockBaseFee uint64
}
