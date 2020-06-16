include Makefile.vars

build: build-ptf
push: push-ptf

build-ptf:
	cd ptf && docker build --build-arg MN_STRATUM_SHA=$(MN_STRATUM_SHA) \
		-t ${PTF_IMG} .

push-ptf:
	# Remember to update Makefile.vars with the new image sha
	docker push ${PTF_IMG}

build-p4pktgen:
	cd p4pktgen && docker build --build-arg B_MODEL_SHA=$(B_MODEL_SHA) \
		-t ${P4PKTGEN_IMG} .

push-p4pktgen:
	docker push ${P4PKTGEN_IMG}

