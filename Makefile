mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
curr_dir := $(patsubst %/,%,$(dir $(mkfile_path)))
curr_dir_sha := $(shell echo -n "$(curr_dir)" | shasum | cut -c1-7)
go_build_name := go-build-${curr_dir_sha}

include util/docker/Makefile.vars

default: build check
build: p4 mapr
check: check-self
check-all: check-self check-dummy check-fabric

.PHONY: ptf mapr

_docker_pull_all:
	docker pull ${P4RT_SH_IMG}@${P4RT_SH_SHA}
	docker tag ${P4RT_SH_IMG}@${P4RT_SH_SHA} ${P4RT_SH_IMG}
	docker pull ${P4C_IMG}@${P4C_SHA}
	docker tag ${P4C_IMG}@${P4C_SHA} ${P4C_IMG}
	docker pull ${MN_STRATUM_IMG}@${MN_STRATUM_SHA}
	docker tag ${MN_STRATUM_IMG}@${MN_STRATUM_SHA} ${MN_STRATUM_IMG}
	docker pull ${PTF_IMG}@${PTF_SHA}
	docker tag ${PTF_IMG}@${PTF_SHA} ${PTF_IMG}
	docker pull ${GNMI_CLI_IMG}@${GNMI_CLI_SHA}
	docker tag ${GNMI_CLI_IMG}@${GNMI_CLI_SHA} ${GNMI_CLI_IMG}
	docker pull ${GOLANG_IMG}

deps: _docker_pull_all

clean:
	-rm -rf p4src/build
	-rm -rf ptf/*.log
	-rm -rf ptf/*.pcap
	-rm -rf mapr/mapr

deep-clean: clean
	-docker container rm ${go_build_name}

p4:
	$(info *** Compiling P4 program...)
	@mkdir -p p4src/build
	@docker run --rm -v ${curr_dir}:/workdir -w /workdir ${P4C_IMG} \
		p4c-bm2-ss --arch v1model -o p4src/build/bmv2.json \
		--p4runtime-files p4src/build/p4info.txt,p4src/build/p4info.bin \
		--Wdisable=unsupported \
		p4src/bng.p4
	@echo "*** P4 program compiled successfully! Output files are in p4src/build"

graph:
	$(info *** Generating P4 program graphs...)
	@mkdir -p p4src/build/graphs
	docker run --rm -v ${curr_dir}:/workdir -w /workdir ${P4C_IMG} \
		p4c-graphs --graphs-dir p4src/build/graphs p4src/bng.p4
	for f in p4src/build/graphs/*.dot; do \
		docker run --rm -v ${curr_dir}:/workdir -w /workdir ${P4C_IMG} \
			dot -Tpdf $${f} > $${f}.pdf; rm -f $${f}; \
	done
	@echo "*** Done! Graph files are in p4src/build/graphs"

_ptf:
	@cd ptf && PTF_DOCKER_IMG=$(PTF_IMG) ./run_tests $(TARGET_NAME) $(TEST)

check-self: TARGET_NAME := self
check-self: _ptf

check-dummy: TARGET_NAME := dummy
check-dummy: _ptf

check-fabric: TARGET_NAME := fabric
check-fabric: _ptf

# Create container once, use it many times to preserve cache.
_go_build_test_container:
	@if ! docker container ls -a --format '{{.Names}}' | grep -q ${go_build_name} ; then \
		docker create -v ${curr_dir}/mapr:/mapr -w /mapr --name ${go_build_name} ${GOLANG_IMG} bash -c "go build && go test ./..."; \
	fi

mapr: _go_build_test_container p4info-go
	$(info *** Building mapr...)
	@docker start -a -i ${go_build_name}

mapr/translate/p4info.go: p4src/build/p4info.txt
mapr/fabric/p4info.go: mapr/p4c-out/fabric/p4info.txt

P4INFO_GO := mapr/translate/p4info.go mapr/fabric/p4info.go
p4info-go: $(P4INFO_GO)
$(P4INFO_GO):
	$(info *** Generating go constants: $< -> $@)
	@docker run -v ${curr_dir}:/tassen -w /tassen \
		--entrypoint ./util/go-gen-p4-const.py $(PTF_IMG) \
		--output $@ --p4info $<
	@docker run -v ${curr_dir}:/tassen -w /tassen \
		${GOLANG_IMG} gofmt -w $@
