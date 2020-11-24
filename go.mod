module github.com/codedropau/openshift-newrelic-synthetics

go 1.14

replace github.com/newrelic/newrelic-client-go => github.com/nickschuch/newrelic-client-go v0.50.1-0.20201124011817-0a6479b171fc

require (
	github.com/newrelic/newrelic-client-go v0.50.0
	github.com/openshift/api v3.9.0+incompatible // indirect
    github.com/openshift/client-go v0.0.0-20201020082437-7737f16e53fc
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.19.4 // indirect
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
)
