This testing app consist of a web server and a client.

The web server is deployed as a service listening on port 666 (on
cluster service network). The service is serviced by one pod running on
controller-1 listening on port 6666 (on cluster pod network).

The client is deployed by daemonset so there is a pod running on each of
the nodes in the cluster. Traffic can be generated from the client pod
to the web service pod by making a http reqest from within the client
pod. The following is an example:

kubectl exec client-ds-69jl7 -- curl web:666

Once such traffic is generated, we can use tcpdump to capture and
analyze the IP packets on the cluster host network interface, to verify
the traffic is encrypted by IPsec for example.

sudo tcpdump -i enp0s8 port 6666

The app can be deployed by applying the yaml file.

kubectl apply -f webserver_client.yml
