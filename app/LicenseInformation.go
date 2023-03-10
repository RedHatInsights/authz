package app

// LicenseInformation returns information about a specific license, including current and total seats.
type LicenseInformation struct {
	// Name of the service the license applies to
	Service string

	// Total available Seats
	SeatsTotal int32

	//Currently available seats
	SeatsAvailable int32
}
