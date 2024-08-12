PREPARE_FILES := $(addprefix debian/,$(shell cat debian/prepare-files))

prepared: prepared-stamp debian/control debian/source.lintian-overrides
prepared-stamp:
	[ -f ext/ext_skel ] && mv ext/ext_skel ext/ext_skel.in
	$(SED) -i -e 's/EXTRA_VERSION=""/EXTRA_VERSION="-$(PHP_DEBIAN_REVISION)"/' configure.in
	./buildconf --force
	touch prepared-stamp

unprepared:
	dh_testdir
	$(SED) -i -e 's/EXTRA_VERSION="-$(PHP_DEBIAN_REVISION)"/EXTRA_VERSION=""/' configure.in
	-[ -f ext/ext_skel.in ] && mv ext/ext_skel.in ext/ext_skel
	rm -f prepared-stamp

prepare-files-stamp:
	# Do this first so we don't overwrite any debhelper-generated files
	#
	# generate versioned files from versionless variants
	for file in $(PREPARE_FILES); do \
	    versionless=$$(echo $$file | $(SED) -e 's/@PHP_VERSION@//g'); \
	    versioned=$$(echo $$file | $(SED) -e 's/@PHP_VERSION@/$(PHP_NAME_VERSION)/g'); \
	    package=$$(echo $$file | $(SED) -e 's,^debian/,,;s/\..*$$//;s/@PHP_VERSION@/$(PHP_NAME_VERSION)/g'); \
	    < $${versionless} $(SED_REPLACEMENT) > $${versioned}; \
	    if [ -x $${versionless} ]; then chmod +x $${versioned}; fi; \
	done
	# generate config snippets and maintscripts for the different sapi implementations
	# from the templates
	for sapi in $(REAL_TARGETS); do \
	    $(SAPI_PACKAGE) \
	    mkdir -p "debian/tmp/usr/lib/php/$(PHP_NAME_VERSION)/sapi/"; \
	    touch "debian/tmp/usr/lib/php/$(PHP_NAME_VERSION)/sapi/$${sapi}"; \
	    for tmpl in postrm prerm postinst dirs install triggers bug-script bug-control; do \
	        < debian/php-sapi.$${tmpl} $(SED_REPLACEMENT) > debian/$${package}.$${tmpl}; \
	        if [ -x debian/php-sapi.$${tmpl} ]; then chmod +x debian/$${package}.$${tmpl}; fi; \
	    done; \
	    < debian/php-sapi.lintian-overrides $(SED_REPLACEMENT) | grep -E "^$${package}" > debian/$${package}.lintian-overrides; \
	done

	for module in $(ext_PACKAGES); do \
	  package=php$(PHP_NAME_VERSION)-$${module}; \
	  extensions=$$(eval echo \$${$${module}_EXTENSIONS}); \
	  description=$$(eval echo \$${$${module}_DESCRIPTION}); \
	  for tmpl in preinst postinst postrm prerm bug-script bug-control triggers dirs substvars lintian-overrides; do \
	      < debian/php-module.$${tmpl}.in \
		$(SED) -e "/\#EXTRA\#/ r debian/$${package}.$${tmpl}.extra" | \
		$(SED) -e "s,@package@,$${package},g"		\
	           -e "s,@extensions@,$${extensions},g"		\
	           -e "s,@module@,$${module},g"			\
	           -e "s|@description@|$${description}|g"	\
	           -e "s,@PHP_VERSION@,$(PHP_NAME_VERSION),g"	\
	           -e "s,@PHP_API@,$${phpapi},g" | \
		$(SED) -e '/\#EXTRA\#/ d' \
	      > debian/$${package}.$${tmpl}; \
	  done; \
	  provides=""; \
	  for dsoname in $${extensions}; do \
	    normalized=$$(echo $${dsoname} | sed -e 's/_/-/g'); \
	    priority=$$(eval echo \$${$${dsoname}_PRIORITY}); \
	    if [ -z "$${priority}" ]; then priority=20; fi; \
	    extension=$$(eval echo \$${$${dsoname}_EXTENSION}); \
	    if [ -z "$${extension}" ]; then extension=extension; fi; \
	    mkdir -p debian/tmp/usr/share/$${package}/$${module}/; \
	    $(SED) -e "s,@extname@,$${module}," \
	           -e "s,@dsoname@,$${dsoname}," \
	           -e "s,@extension@,$${extension}," \
	           -e "s,@priority@,$${priority}," \
	      < debian/php-module.ini.in \
	      > debian/tmp/usr/share/$${package}/$${module}/$${dsoname}.ini; \
	    echo "usr/lib/php/*/$${dsoname}.so" >> debian/$${package}.install; \
	    echo "usr/share/$${package}/$${module}/$${dsoname}.ini" >> debian/$${package}.install; \
	    if [ "$${normalized}" != "gettext" ]; then provides="php-$${normalized}, $${provides}"; fi; \
	    if [ "$${module}" != "$${normalized}" ]; then provides="php$(PHP_NAME_VERSION)-$${normalized}, $${provides}"; fi; \
	  done; \
	  echo "php-$${module}:Provides=$${provides}" >> debian/$${package}.substvars; \
	done
	touch prepare-files-stamp

remove-files-stamp:
	# get rid of dreaded libtool files
	find debian/tmp/ -name '*.la' -delete
	# get rid of static versions of PHP modules (WTF?)
	rm -f debian/tmp/usr/lib/php/$(PHP_ZEND_VERSION)/*.a

	rm -rf \
	  debian/tmp/.filemap \
	  debian/tmp/.channels \
	  debian/tmp/.lock \
	  debian/tmp/.depdb* \
	  debian/tmp/usr/bin/pear* \
	  debian/tmp/usr/bin/pecl* \
	  debian/tmp/usr/share/php/.filemap \
	  debian/tmp/usr/share/php/.lock \
	  debian/tmp/usr/share/php/.depdb* \
	  debian/tmp/usr/share/php/*.php \
	  debian/tmp/usr/share/php/.registry/ \
	  debian/tmp/usr/share/php/.channels/ \
	  debian/tmp/usr/share/php/doc/ \
	  debian/tmp/usr/share/php/Archive/ \
	  debian/tmp/usr/share/php/Console/ \
	  debian/tmp/usr/share/php/Structures/ \
	  debian/tmp/usr/share/php/test/ \
	  debian/tmp/usr/share/php/XML/ \
	  debian/tmp/usr/share/php/OS/ \
	  debian/tmp/usr/share/php/PEAR/ \
	  debian/tmp/usr/share/php/data/ \
	  debian/tmp/etc/pear.conf
	# shipping duplicate files from other packages is hell for security audits
	rm -f \
	  debian/tmp$(PHPIZE_BUILDDIR)/config.guess \
	  debian/tmp$(PHPIZE_BUILDDIR)/config.sub \
	  debian/tmp$(PHPIZE_BUILDDIR)/libtool.m4 \
	  debian/tmp$(PHPIZE_BUILDDIR)/pkg.m4 \
	  debian/tmp$(PHPIZE_BUILDDIR)/ltmain.sh \
	  debian/tmp$(PHPIZE_BUILDDIR)/shtool
	touch remove-files-stamp

debian/control: debian/control.in debian/rules debian/changelog debian/source.lintian-overrides debian/rules.d/* debian/php-module.control.in
	$(SED) -e "s/@PHP_VERSION@/$(PHP_NAME_VERSION)/g" -e "s/@BUILT_USING@/$(BUILT_USING)/g" >$@ <$<
	for ext in $(ext_PACKAGES); do \
	  package=php$(PHP_NAME_VERSION)-$${ext}; \
	  description=$$(eval echo \$${$${ext}_DESCRIPTION}); \
	  echo >>$@; \
	  $(SED) -e "s|@ext@|$${ext}|" -e "s|@package@|$${package}|" -e "s|@description@|$${description}|" >>$@ <debian/php-module.control.in; \
	done
	mkdir -p debian/tests
	for f in debian/tests.in/*; do \
	  t=$$(basename $${f}); \
	  < debian/tests.in/$${t} $(SED_REPLACEMENT) > debian/tests/$${t}; \
	done

debian/source.lintian-overrides: debian/source.lintian-overrides.in debian/rules debian/changelog
	$(SED) -e "s/@PHP_VERSION@/$(PHP_NAME_VERSION)/g" >$@ <$<

prepare: debian/control debian/source.lintian-overrides
