package resources

import (
	"context"

	"github.com/go-logr/logr"
	storagev1 "k8s.io/api/storage/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StorageClass reconciles a StorageClass object.
func StorageClass(ctx context.Context, r client.Client, storageClass *storagev1.StorageClass, log logr.Logger) error {
	foundStorageClass := &storagev1.StorageClass{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: storageClass.Name}, foundStorageClass); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating StorageClass", "name", storageClass.Name)
			if err = r.Create(ctx, storageClass); err != nil {
				log.Error(err, "Unable to create StorageClass")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Error getting StorageClass")
			return err
		}
	}
	if !justCreated && CopyStorageClass(storageClass, foundStorageClass) {
		log.Info("Updating StorageClass", "name", storageClass.Name)
		if err := r.Update(ctx, foundStorageClass); err != nil {
			log.Error(err, "Unable to update StorageClass")
			return err
		}
	}

	return nil
}

// CopyStorageClass copies the owned fields from one StorageClass to another
func CopyStorageClass(from, to *storagev1.StorageClass) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	return requireUpdate
}
