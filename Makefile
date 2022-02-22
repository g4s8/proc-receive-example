TESTDIR = _test

.PHONY: all
all: test

.PHONY: build
# build binary for proc-receive hook
build: proc-receive

proc-receive: main.go
	go build -o proc-receive main.go

.PHONY: test-deploy
# deploy proc-receive hook to test server
test-deploy: proc-receive $(TESTDIR)/server
	cp proc-receive $(TESTDIR)/server/hooks/proc-receive

.PHONY: test
# test client and server communication
test: test-deploy prepare-test
	cd $(TESTDIR)/client && \
		echo $(shell date) >> date.txt && \
		git add date.txt && \
		git commit -m "update date" && \
		git push -v origin master

.PHONY: prepare-test
# prepare test repositories
prepare-test: $(TESTDIR)/server $(TESTDIR)/client


# prepare test server dir
$(TESTDIR)/server:
	mkdir -p $(TESTDIR)/server
	git --git-dir $(TESTDIR)/server init --bare
	git --git-dir $(TESTDIR)/server config --local receive.advertisePushOPtions true
	git --git-dir $(TESTDIR)/server config --local receive.procReceiveRefs refs/

# prepare test client dir
$(TESTDIR)/client:
	mkdir -p $(TESTDIR)/client
	git init $(TESTDIR)/client
	git --git-dir $(TESTDIR)/client/.git remote add origin $(PWD)/$(TESTDIR)/server

.PHONY: clean-test
clean-test:
	rm -fr $(TESTDIR)

.PHONY: clean
clean:
	rm -f proc-receive
