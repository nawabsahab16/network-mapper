package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"strings"

	"github.com/otterize/network-mapper/src/mapper/pkg/config"
	"github.com/otterize/network-mapper/src/mapper/pkg/graph/generated"
	"github.com/otterize/network-mapper/src/mapper/pkg/graph/model"
	"github.com/otterize/network-mapper/src/mapper/pkg/intentsstore"
	"github.com/otterize/network-mapper/src/mapper/pkg/kubefinder"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

func (r *mutationResolver) ResetCapture(ctx context.Context) (bool, error) {
	logrus.Info("Resetting stored intents")
	r.intentsHolder.Reset()
	return true, nil
}

func (r *mutationResolver) ReportCaptureResults(ctx context.Context, results model.CaptureResults) (bool, error) {
	for _, captureItem := range results.Results {
		srcPod, err := r.kubeFinder.ResolveIpToPod(ctx, captureItem.SrcIP)
		if err != nil {
			if errors.Is(err, kubefinder.ErrFoundMoreThanOnePod) {
				logrus.WithError(err).Debugf("Ip %s belongs to more than one pod, ignoring", captureItem.SrcIP)
			} else {
				logrus.WithError(err).Debugf("Could not resolve %s to pod", captureItem.SrcIP)
			}
			continue
		}
		srcService, err := r.serviceIdResolver.ResolvePodToServiceIdentity(ctx, srcPod)
		if err != nil {
			logrus.WithError(err).Debugf("Could not resolve pod %s to identity", srcPod.Name)
			continue
		}
		for _, dest := range captureItem.Destinations {
			destAddress := dest.Destination
			if !strings.HasSuffix(destAddress, viper.GetString(config.ClusterDomainKey)) {
				// not a k8s service, ignore
				continue
			}
			ips, err := r.kubeFinder.ResolveServiceAddressToIps(ctx, destAddress)
			if err != nil {
				logrus.WithError(err).Warningf("Could not resolve service address %s", dest)
				continue
			}
			if len(ips) == 0 {
				logrus.Debugf("Service address %s is currently not backed by any pod, ignoring", dest)
				continue
			}
			destPod, err := r.kubeFinder.ResolveIpToPod(ctx, ips[0])
			if err != nil {
				if errors.Is(err, kubefinder.ErrFoundMoreThanOnePod) {
					logrus.WithError(err).Debugf("Ip %s belongs to more than one pod, ignoring", ips[0])
				} else {
					logrus.WithError(err).Debugf("Could not resolve %s to pod", ips[0])
				}
				continue
			}
			dstService, err := r.serviceIdResolver.ResolvePodToServiceIdentity(ctx, destPod)
			if err != nil {
				logrus.WithError(err).Debugf("Could not resolve pod %s to identity", destPod.Name)
				continue
			}

			srcSvcIdentity := model.OtterizeServiceIdentity{Name: srcService.Name, Namespace: srcPod.Namespace, Labels: podLabelsToOtterizeLabels(srcPod)}
			dstSvcIdentity := model.OtterizeServiceIdentity{Name: dstService.Name, Namespace: destPod.Namespace, Labels: podLabelsToOtterizeLabels(destPod)}
			if srcService.OwnerObject != nil {
				srcSvcIdentity.PodOwnerKind = model.GroupVersionKindFromKubeGVK(srcService.OwnerObject.GetObjectKind().GroupVersionKind())
			}

			if dstService.OwnerObject != nil {
				dstSvcIdentity.PodOwnerKind = model.GroupVersionKindFromKubeGVK(dstService.OwnerObject.GetObjectKind().GroupVersionKind())
			}

			intent := model.Intent{
				Client: &srcSvcIdentity,
				Server: &dstSvcIdentity,
			}

			r.intentsHolder.AddIntent(
				dest.LastSeen,
				intent,
			)
		}
	}
	return true, nil
}

func (r *mutationResolver) ReportSocketScanResults(ctx context.Context, results model.SocketScanResults) (bool, error) {
	for _, socketScanItem := range results.Results {
		srcPod, err := r.kubeFinder.ResolveIpToPod(ctx, socketScanItem.SrcIP)
		if err != nil {
			if errors.Is(err, kubefinder.ErrFoundMoreThanOnePod) {
				logrus.WithError(err).Debugf("Ip %s belongs to more than one pod, ignoring", socketScanItem.SrcIP)
			} else {
				logrus.WithError(err).Debugf("Could not resolve %s to pod", socketScanItem.SrcIP)
			}
			continue
		}
		srcService, err := r.serviceIdResolver.ResolvePodToServiceIdentity(ctx, srcPod)
		if err != nil {
			logrus.WithError(err).Debugf("Could not resolve pod %s to identity", srcPod.Name)
			continue
		}
		for _, destIp := range socketScanItem.DestIps {
			destPod, err := r.kubeFinder.ResolveIpToPod(ctx, destIp.Destination)
			if err != nil {
				if errors.Is(err, kubefinder.ErrFoundMoreThanOnePod) {
					logrus.WithError(err).Debugf("Ip %s belongs to more than one pod, ignoring", destIp)
				} else {
					logrus.WithError(err).Debugf("Could not resolve %s to pod", destIp)
				}
				continue
			}
			dstService, err := r.serviceIdResolver.ResolvePodToServiceIdentity(ctx, destPod)
			if err != nil {
				logrus.WithError(err).Debugf("Could not resolve pod %s to identity", destPod.Name)
				continue
			}

			srcSvcIdentity := model.OtterizeServiceIdentity{Name: srcService.Name, Namespace: srcPod.Namespace, Labels: podLabelsToOtterizeLabels(srcPod)}
			dstSvcIdentity := model.OtterizeServiceIdentity{Name: dstService.Name, Namespace: destPod.Namespace, Labels: podLabelsToOtterizeLabels(destPod)}
			if srcService.OwnerObject != nil {
				srcSvcIdentity.PodOwnerKind = model.GroupVersionKindFromKubeGVK(srcService.OwnerObject.GetObjectKind().GroupVersionKind())
			}

			if dstService.OwnerObject != nil {
				dstSvcIdentity.PodOwnerKind = model.GroupVersionKindFromKubeGVK(dstService.OwnerObject.GetObjectKind().GroupVersionKind())
			}

			intent := model.Intent{
				Client: &srcSvcIdentity,
				Server: &dstSvcIdentity,
			}

			r.intentsHolder.AddIntent(
				destIp.LastSeen,
				intent,
			)
		}
	}
	return true, nil
}

func (r *mutationResolver) ReportKafkaMapperResults(ctx context.Context, results model.KafkaMapperResults) (bool, error) {
	for _, result := range results.Results {
		srcPod, err := r.kubeFinder.ResolveIpToPod(ctx, result.SrcIP)
		if err != nil {
			if errors.Is(err, kubefinder.ErrFoundMoreThanOnePod) {
				logrus.WithError(err).Debugf("Ip %s belongs to more than one pod, ignoring", result.SrcIP)
			} else {
				logrus.WithError(err).Debugf("Could not resolve %s to pod", result.SrcIP)
			}
			continue
		}
		srcService, err := r.serviceIdResolver.ResolvePodToServiceIdentity(ctx, srcPod)
		if err != nil {
			logrus.WithError(err).Debugf("Could not resolve pod %s to identity", srcPod.Name)
			continue
		}

		srcSvcIdentity := model.OtterizeServiceIdentity{Name: srcService.Name, Namespace: srcPod.Namespace, Labels: podLabelsToOtterizeLabels(srcPod)}

		dstPod, err := r.kubeFinder.ResolvePodByName(ctx, result.ServerPodName, result.ServerNamespace)
		if err != nil {
			logrus.WithError(err).Debugf("Could not resolve pod %s to identity", result.ServerPodName)
			continue
		}
		dstService, err := r.serviceIdResolver.ResolvePodToServiceIdentity(ctx, dstPod)
		if err != nil {
			logrus.WithError(err).Debugf("Could not resolve pod %s to identity", dstPod.Name)
			continue
		}
		dstSvcIdentity := model.OtterizeServiceIdentity{Name: dstService.Name, Namespace: dstPod.Namespace, Labels: podLabelsToOtterizeLabels(dstPod)}

		operation, err := model.KafkaOpFromText(result.Operation)
		if err != nil {
			logrus.WithError(err).Debugf("Could not resolve kafka operation %s", result.Operation)
			continue
		}

		intent := model.Intent{
			Client: &srcSvcIdentity,
			Server: &dstSvcIdentity,
			Type:   lo.ToPtr(model.IntentTypeKafka),
			KafkaTopics: []model.KafkaConfig{
				{
					Name:       result.Topic,
					Operations: []model.KafkaOperation{operation},
				},
			},
		}

		r.intentsHolder.AddIntent(
			result.LastSeen,
			intent,
		)
	}

	return true, nil
}

func (r *mutationResolver) ReportIstioConnectionResults(ctx context.Context, results model.IstioConnectionResults) (bool, error) {
	for _, result := range results.Results {
		r.intentsHolder.AddIntent(result.LastSeen, model.Intent{
			Client: &model.OtterizeServiceIdentity{
				Name:      result.SrcWorkload,
				Namespace: result.SrcWorkloadNamespace,
			},
			Server: &model.OtterizeServiceIdentity{
				Name:      result.DstWorkload,
				Namespace: result.DstWorkloadNamespace,
			},
			Type: lo.ToPtr(model.IntentTypeHTTP),
			HTTPResources: lo.Map(result.RequestPaths, func(path string, i int) model.HTTPResource {
				return model.HTTPResource{Path: path}
			}),
		})
	}
	return true, nil
}

func (r *queryResolver) ServiceIntents(ctx context.Context, namespaces []string, includeLabels []string, includeAllLabels *bool) ([]model.ServiceIntents, error) {
	shouldIncludeAllLabels := false
	if includeAllLabels != nil && *includeAllLabels {
		shouldIncludeAllLabels = true
	}
	discoveredIntents := r.intentsHolder.GetIntents(namespaces, includeLabels, shouldIncludeAllLabels)
	intentsBySource := intentsstore.GroupIntentsBySource(discoveredIntents)

	// sorting by service name so results are more consistent
	slices.SortFunc(intentsBySource, func(intentsa, intentsb model.ServiceIntents) bool {
		return intentsa.Client.AsNamespacedName().String() < intentsb.Client.AsNamespacedName().String()
	})

	for _, intents := range intentsBySource {
		slices.SortFunc(intents.Intents, func(desta, destb model.OtterizeServiceIdentity) bool {
			return desta.AsNamespacedName().String() < destb.AsNamespacedName().String()
		})
	}

	return intentsBySource, nil
}

func (r *queryResolver) Intents(ctx context.Context, namespaces []string, includeLabels []string, includeAllLabels *bool) ([]model.Intent, error) {
	shouldIncludeAllLabels := false
	if includeAllLabels != nil && *includeAllLabels {
		shouldIncludeAllLabels = true
	}
	timestampedIntents := r.intentsHolder.GetIntents(namespaces, includeLabels, shouldIncludeAllLabels)
	intents := lo.Map(timestampedIntents, func(timestampedIntent intentsstore.TimestampedIntent, _ int) model.Intent {
		return timestampedIntent.Intent
	})

	// sort by service names for consistent ordering
	slices.SortFunc(intents, func(intenta, intentb model.Intent) bool {
		clienta, clientb := intenta.Client.AsNamespacedName(), intentb.Client.AsNamespacedName()
		servera, serverb := intenta.Server.AsNamespacedName(), intentb.Server.AsNamespacedName()

		if clienta != clientb {
			return clienta.String() < clientb.String()
		}

		return servera.String() < serverb.String()
	})

	return intents, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
