package timeline

/**
* The transport interface to be implemented.
* @author rnojiri
**/

// Transport - the implementation type to send a event
type Transport interface {

	// Send - send a new point
	PointChannel() chan<- interface{}

	// ConfigureBackend - configures the backend
	ConfigureBackend(backend *Backend) error

	// Start - starts this transport
	Start() error

	// Close - closes this transport
	Close()
}
