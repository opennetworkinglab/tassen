include ../docker/Makefile.vars

build-p4:
	$(info *** Compiling P4 program for p4pktgen...)
	@mkdir -p p4build
	@docker run --rm -v $(shell pwd)/../../:/workdir -w /workdir ${P4C_IMG} \
		p4c-bm2-ss --arch v1model -o util/p4pktgen/p4build/bmv2.json \
		--p4runtime-files util/p4pktgen/p4build/p4info.txt,util/p4pktgen/p4build/p4info.bin \
		--Wdisable=unsupported -DFOR_P4PKTGEN\
		p4src/bng.p4
	@echo "*** P4 program for p4pktgen compiled successfully! Output files are in util/p4pktgen/p4build"

p4pktgen: build-p4
	$(info *** Run p4pktgen against bng.p4...)
	@docker run --rm -v $(shell pwd):/tassen_p4pktgen -ti -w /tassen_p4pktgen ${P4PKTGEN_IMG} \
		/run_p4pktgen.sh ./p4build/bmv2.json > ./log.txt 2>&1
	@echo "*** Output files are in util/p4pktgen"
