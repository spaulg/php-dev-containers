override_dh_auto_configure-indep:

override_dh_auto_configure-arch: prepared
	for target in $(TARGETS); do \
	  dh_auto_configure --builddirectory $${target}-build $(PARALLEL) -- $$(eval echo \$${$${target}_config}); \
	done

override_dh_auto_build-indep:

override_dh_auto_build-arch:
	for target in $(TARGETS); do \
	  dh_auto_build --builddirectory $${target}-build $(PARALLEL); \
	done

override_dh_auto_test-indep:

override_dh_auto_test-arch:
ifeq (yes,$(RUN_TESTS))
	mkdir -p temp_session_store
	extensions=""; \
	for f in $(CURDIR)/ext-build/modules/*.so; do \
	    ext=`basename "$$f"`; \
	    test -d "$(CURDIR)/ext/$${ext%.so}/tests" || continue; \
	    test "$$ext" != "imap.so" || continue; \
	    test "$$ext" != "interbase.so" || continue; \
	    test "$$ext" != "ldap.so" || continue; \
	    test "$$ext" != "odbc.so" || continue; \
	    test "$$ext" != "pgsql.so" || continue; \
	    test "$$ext" != "pdo_dblib.so" || continue; \
	    test "$$ext" != "pdo_firebird.so" || continue; \
	    test "$$ext" != "pdo_odbc.so" || continue; \
	    test "$$ext" != "pdo_pgsql.so" || continue; \
	    test "$$ext" != "snmp.so" || continue; \
	    test "$$ext" != "opcache.so" || continue; \
	    test "$$ext" != "mysqlnd.so" || continue; \
	    test "$$ext" != "mysqli.so" || continue; \
	    test "$$ext" != "pdo_mysql.so" || continue; \
	    test "$$ext" != "wddx.so" || continue; \
	    extensions="$$extensions -d extension=$$ext"; \
	done; \
	[ "$$extensions" ] || { echo "extensions list is empty"; exit 1; }; \
	env NO_INTERACTION=1 \
	    TEST_PHP_CGI_EXECUTABLE=$(CURDIR)/cgi-build/sapi/cgi/php-cgi \
	    TEST_PHP_EXECUTABLE=$(CURDIR)/cli-build/sapi/cli/php \
	$(CURDIR)/cli-build/sapi/cli/php run-tests.php \
		-n \
		-d extension_dir=$(CURDIR)/ext-build/modules/ \
		$$extensions | \
	tee test-results.txt
	rm -rf temp_session_store
	@for test in `find . -name '*.log' -a '!' -name 'config.log' -a '!' -name 'bootstrap.log' -a '!' -name 'run.log'`; do \
	    echo; \
	    echo -n "$${test#./}:"; \
	    cat $$test; \
	    echo; \
	done | tee -a test-results.txt
else
	echo 'nocheck found in DEB_BUILD_OPTIONS or unsupported architecture' | tee test-results.txt
endif
