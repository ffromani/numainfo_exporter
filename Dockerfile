FROM registry.access.redhat.com/ubi8/ubi-minimal

#USER nobody
ADD _output/bin/numainfo_exporter /bin/numainfo_exporter
