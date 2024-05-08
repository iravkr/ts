CCO-MON XAPP

This xApp implmements the following functionalities: 
- Sending TX power change requests for specific cells to the ran simulator using RC service model.
- Collecting the rsrp measurements from the serving cell and neighboring cells for each UE in the simulation scenario using MHO service model. 

The communication between the xApp and the northbound client application is implemented using grpc APIs. client.go application sends tx power change 
requests and receives rsrp reports in a simultaneous manner. The client then dumps the data (rsrp reports and tx power levels) into a csv file: grpc-client/cco-mon.csv

Notes: 
- KPImon Xapp should not be running with this CCO-MON xApp, disable it by using "--set import.onos-kpimon.enabled=false " in the MakefileVar.mk HELM_ARGS
- Copy the helm chart folder for this xApp to the directory: sdran-in-a-box/workspace/helm-charts/sdran-helm-charts
- Deploy the attached ran-simulator source code as some changes were made there as well
- The grpc client should run on the machine outside the onos K8s cluster using: "cco-mon/pkg/grpc-client$ go run . "
- Replace the IP address in the grpc.Dial in grpc-client/client.go with the IP address of the k8s node 

kubectl get nodes -o wide

Running Locally
kubectl port-forward svc/onos-topo 9090:5150 -n riab
kubectl port-forward svc/onos-e2t 9091:5150 -n riab
cco-mon server: go run ./cmd/cco-mon/ -topoAddress localhost -topoPort 9090 -e2tAddress localhost e2tPort 9091 -e2tEndpoint localhost:9091
client: cd pkg/grpc-client/ && go run ./

## Caveats in the sdk
onos-ric-sdk using hardcoded string (localhost:5151) to point to e2t service
https://github.com/onosproject/onos-ric-sdk-go/blob/35465e07efff60d96b0f38dbf65801f710f28999/pkg/e2/v1beta1/node.go#L125
This is updated (vendor/github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1/node.go:L125)