PCNTL_FUNCTIONS := $(shell < ext/pcntl/php_pcntl.h $(SED) -ne "/^PHP_FUNCTION/ s/PHP_FUNCTION(\(.*\));/\1/;t end;d;:end p" | tr '\n' ',')

override_dh_installdirs: prepare-files-stamp
	dh_installdirs

override_dh_auto_install:
	for target in $(TARGETS); do \
	  dh_auto_install --builddirectory $${target}-build $(PARALLEL) -- install INSTALL_ROOT=$(CURDIR)/debian/tmp; \
	done

override_dh_install-arch: remove-files-stamp prepare-fpm-pools
# Rename Apache2 SAPI
	mv -u debian/tmp/usr/lib/apache2/modules/libphp$(PHP_MAJOR_VERSION).so debian/tmp/usr/lib/apache2/modules/libphp$(PHP_NAME_VERSION).so
# Rename embed SAPI
	mv -u debian/tmp/usr/lib/php/libphp$(PHP_MAJOR_VERSION).so debian/tmp/usr/lib/libphp$(PHP_NAME_VERSION).so
# Install extra CGI-BIN
	install -d -m 755 debian/tmp/usr/lib/cgi-bin/
	ln debian/tmp/usr/bin/php-cgi$(PHP_NAME_VERSION) debian/tmp/usr/lib/cgi-bin/php$(PHP_NAME_VERSION)

	dh_install --fail-missing

# Install a helper script for checking PHP FPM configuration
	mkdir -p debian/$(PHP_FPM)/usr/lib/php/
	install -m 755 debian/$(PHP_FPM)-reopenlogs debian/$(PHP_FPM)/usr/lib/php/

	# sanitize php.ini files
	mkdir -p debian/$(PHP_COMMON)/usr/lib/php/$(PHP_NAME_VERSION)/
	cat php.ini-production | tr "\t" " " | \
	$(SED) -e'/session.gc_probability =/ s/1/0/g;' \
	    -e'/disable_functions =/ s/$$/ $(PCNTL_FUNCTIONS)/g;' \
	    -e'/expose_php =/ s/On/Off/g;' \
	  > debian/$(PHP_COMMON)/usr/lib/php/$(PHP_NAME_VERSION)/php.ini-production

	cat php.ini-production | tr "\t" " " | \
	$(SED) -e'/memory_limit =/ s/128M/-1/g;' \
	    -e'/session.gc_probability =/ s/1/0/g' \
	  > debian/$(PHP_COMMON)/usr/lib/php/$(PHP_NAME_VERSION)/php.ini-production.cli

	cat php.ini-development | tr "\t" " " | \
	$(SED) -e'/session.gc_probability =/ s/1/0/g;' \
	    -e'/disable_functions =/ s/$$/ $(PCNTL_FUNCTIONS)/g;' \
	  > debian/$(PHP_COMMON)/usr/lib/php/$(PHP_NAME_VERSION)/php.ini-development

ifeq (yes,$(RUN_TESTS))
	mkdir -p debian/$(PHP_COMMON)/usr/share/doc/$(PHP_COMMON)/
endif

	$(SED) -i -e's@-ffile-prefix-map=[^ ]*[ ]*@@g' \
		-e's@-fdebug-prefix-map=[^ ]*[ ]*@@g' \
		-e's@$(CURDIR)@/tmp/buildd/nonexistent@g' \
		debian/$(PHP_DEV)/usr/include/php/*/main/build-defs.h \
		debian/$(PHP_DEV)/usr/bin/php-config$(PHP_NAME_VERSION)

override_dh_installdocs-indep:
	dh_installdocs -i

override_dh_installdocs-arch:
	dh_installdocs -p$(PHP_COMMON)
	dh_installdocs -a --remaining-packages --link-doc=$(PHP_COMMON)

override_dh_installchangelogs-arch:
	dh_installchangelogs -a -p$(PHP_COMMON) NEWS

override_dh_installchangelogs-indep:
	dh_installchangelogs -i NEWS

override_dh_installinit:
	dh_installinit --restart-after-upgrade

override_dh_systemd_start:
	dh_systemd_start --restart-after-upgrade

override_dh_apache2:
	for sapi in apache2 cgi fpm; do \
	    $(SAPI_PACKAGE) \
	    < debian/$${versionless}.apache2 $(SED_REPLACEMENT) > debian/$${package}.apache2; \
	done
	dh_apache2 --conditional=php_enable

override_dh_compress:
	dh_compress -Xphp.ini

override_dh_strip:
	dh_strip --dbgsym-migration='php$(PHP_NAME_VERSION)-dbg' || dh_strip

override_dh_makeshlibs-arch:
	dh_makeshlibs -a -p$(PHP_LIBEMBED) -V '$(PHP_LIBEMBED) (>= $(PHP_MAJOR_VERSION).$(PHP_MINOR_VERSION))'

#override_dh_gencontrol-arch:
#	# Bail-out if PHPAPI has changed
#	stored=$$(cat debian/phpapi); \
#	for sapi in $(REAL_TARGETS); do \
#	    $(SAPI_PACKAGE) \
#	    $${sapi}-build/sapi/cli/php -n -r '$(BUILTIN_EXTENSION_CHECK)' \
#	      >> debian/$${package}.substvars; \
#	    phpapi=$$(sh $${sapi}-build/scripts/php-config --phpapi); \
#	    if [ "$${phpapi}" != "$${stored}" ]; then \
#	        echo "PHPAPI has changed from $${stored} to $${phpapi}, please modify debian/phpapi"; \
#	        exit 1; \
#	    fi; \
#	    echo "php:Provides=phpapi-$${phpapi}" >> debian/$${package}.substvars; \
#	done; \
#	if dpkg --compare-versions $(LIBTOOL_VERSION) gt 2.4.6-0.1~; then \
#	    echo "libtool:Depends=libtool (>= 2.4.6-0.1~)" >> debian/php$(PHP_NAME_VERSION)-dev.substvars; \
#	else \
#	    echo "libtool:Depends=libtool" >> debian/php$(PHP_NAME_VERSION)-dev.substvars; \
#	fi
#	dh_gencontrol -a

