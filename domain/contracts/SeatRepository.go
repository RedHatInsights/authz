package contracts

const j = "abc"

// SeatRepository
// - AddSeats -> adds unseated/unlicensed users of an org to a seatlicense for a specific service
// => calls CheckPermission & writeRelationships under the hood
// - RemoveSeats -> removes seated/licensed users of an org from a seatlicense for a specific service
// => calls CheckPermission & (batch) deleteRelationship under the hood
//TODO: Maybe this should not exist, as I understand it the actual "SeatLicense" is the aggregate root, so thar repo should contain everything
