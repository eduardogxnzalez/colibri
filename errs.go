package colibri

import (
	"encoding/json"
	"strconv"
	"sync"
)

// AddError adds an error to the existing error set.
// If errs or err is null or the key is empty, no operation is performed.
// If errs is not of type *Err, a new error of type *Err is returned
// and the original error is stored with the key "#".
func AddError(errs error, key string, err error) error {
	if (errs == nil) && ((key == "") || (err == nil)) {
		return nil
	}

	e, ok := errs.(*Errs)
	if ok {
		return e.Add(key, err)
	}

	e = &Errs{}
	if errs != nil {
		e.Add("#", errs)
	}
	return e.Add(key, err)
}

// Errs is a structure that stores and manages errors.
type Errs struct {
	rw   sync.RWMutex
	data map[string]error
}

// Add adds an error to the error set.
// If the key or error is null, no operation is performed.
// If there is already an error stored with the same key,
// the error is stored with the key + # + key number.
// Returns a pointer to the updated error structure.
func (errs *Errs) Add(key string, err error) *Errs {
	if (key == "") || (err == nil) {
		return errs
	}

	errs.rw.Lock()
	if errs.data == nil {
		errs.data = make(map[string]error)
	}

	keyAux := key + "#"
	for i := 1; ; i++ {
		if _, ok := errs.data[key]; !ok {
			break
		}
		key = keyAux + strconv.Itoa(i)
	}

	errs.data[key] = err
	errs.rw.Unlock()
	return errs
}

// Get returns the error associated with a key and
// a boolean indicating whether the key exists.
// If the key does not exist, a null error and false are returned.
func (errs *Errs) Get(key string) (err error, ok bool) {
	errs.rw.RLock()
	err, ok = errs.data[key]
	errs.rw.RUnlock()
	return err, ok
}

// Error returns a string representation of errors stored in JSON format.
func (errs *Errs) Error() string {
	b, _ := errs.MarshalJSON()
	return string(b)
}

// MarshalJSON returns the JSON representation of the stored errors.
func (errs *Errs) MarshalJSON() ([]byte, error) {
	errs.rw.Lock()
	defer errs.rw.Unlock()

	errsMap := make(map[string]any, len(errs.data))
	for key, err := range errs.data {
		if e, ok := err.(json.Marshaler); ok {
			errsMap[key] = e
			continue
		}
		errsMap[key] = err.Error()
	}
	return json.Marshal(errsMap)
}
