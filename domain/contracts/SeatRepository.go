package contracts

const j = "abc"

// SeatRepository
// - AddSeats -> adds unseated/unlicensed users of an org to a seatlicense for a specific service
// => calls CheckPermission & writeRelationships under the hood
// - RemoveSeats -> removes seated/licensed users of an org from a seatlicense for a specific service
// => calls CheckPermission & (batch) deleteRelationship under the hood
