// Code generated by neuron-generator. DO NOT EDIT.
// This file was generated at:
// Mon, 27 Jul 2020 14:19:08 +0200

package account

type _collectionInitializer func() error

var _collectionInitializers []_collectionInitializer

// Neuron_InitializeCollections collections for provided controller.
func Neuron_InitializeCollections() (err error) {
	for _, initializer := range _collectionInitializers {
		if err = initializer(); err != nil {
			return err
		}
	}
	return nil
}
