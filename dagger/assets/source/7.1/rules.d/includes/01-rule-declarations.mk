# Uncomment this to turn on verbose mode.
export DH_VERBOSE=1

# This has to be exported to make some magic below work.
export DH_OPTIONS

# Hardcode correct absolute path to sed
export ac_cv_path_SED = $(shell command -v sed)

SANE_ARCHS := amd64 i386 arm64

# Enable parallel builds
PARALLEL=--parallel

# Enable this for debugging the sed scripts
#SED=$(CURDIR)/debian/sedsed
export SED := /bin/sed

# Make the shell scripts fail after first failed command (important for SAPI loops)
SHELL := /bin/sh -e

# enable dpkg build flags
export DEB_BUILD_MAINT_OPTIONS = hardening=+all optimize=-lto
DPKG_EXPORT_BUILDFLAGS = 1
include /usr/share/dpkg/default.mk

export DEB_HOST_MULTIARCH
PHP_SOURCE_VERSION   := $(DEB_VERSION)
PHP_UPSTREAM_VERSION := $(DEB_VERSION_UPSTREAM)
PHP_DEBIAN_REVISION  := $(shell echo $(PHP_SOURCE_VERSION) | $(SED) -e 's/.*-//')
PHP_DFSG_VERSION     := $(shell echo $(PHP_UPSTREAM_VERSION) | $(SED) -e 's/+.*//')
PHP_MAJOR_VERSION    := $(shell echo $(PHP_DFSG_VERSION) | awk -F. '{print $$1}')
PHP_MINOR_VERSION    := $(shell echo $(PHP_DFSG_VERSION) | awk -F. '{print $$2}')
PHP_RELEASE_VERSION  := $(shell echo $(PHP_DFSG_VERSION) | awk -F. '{print $$3}')

# Enable ZTS build if $(DEB_SOURCE) ends with -zts
ZTS=$(shell echo $(DEB_SOURCE) | sed 's/php$(PHP_MAJOR_VERSION).$(PHP_MINOR_VERSION)//')
ifeq ($(ZTS),-zts)
$(warning Enabling ZTS build)
CONFIGURE_ZTS        := --enable-maintainer-zts
endif

PHP_NAME_VERSION     := $(PHP_MAJOR_VERSION).$(PHP_MINOR_VERSION)$(ZTS)
PHP_ZEND_VERSION     := $(shell $(SED) -ne 's/\#define ZEND_MODULE_API_NO //p' Zend/zend_modules.h)$(ZTS)

ifneq ($(DEB_SOURCE),php$(PHP_NAME_VERSION))
$(error $(DEB_SOURCE) != php$(PHP_NAME_VERSION))
endif
REAL_TARGETS         := apache2 phpdbg embed fpm cgi cli
EXTRA_TARGETS        := ext
TARGETS              := $(EXTRA_TARGETS) $(REAL_TARGETS)

# Special package names
PHP_PHP      := php$(PHP_NAME_VERSION)
PHP_COMMON   := php$(PHP_NAME_VERSION)-common
PHP_FPM      := php$(PHP_NAME_VERSION)-fpm
PHP_LIBEMBED := libphp$(PHP_NAME_VERSION)-embed
PHP_DEV      := php$(PHP_NAME_VERSION)-dev
PHP_APACHE2  := libapache2-mod-php$(PHP_NAME_VERSION)
PHP_CGI      := php$(PHP_NAME_VERSION)-cgi
PHP_CLI      := php$(PHP_NAME_VERSION)-cli
PHP_PHPDBG   := php$(PHP_NAME_VERSION)-phpdbg

# Generic commands

SED_VARIABLES := \
	$(SED) -e "s,@sapi@,$${sapi},g"				|\
	$(SED) -e "s,@package@,$${package},g"			|\
	$(SED) -e "s,@extensions@,$${extensions},g"		|\
	$(SED) -e "s,@module@,$${module},g"			|\
	$(SED) -e "s,@extdir@,$${extdir},g"			|\
	$(SED) -e "s,@priority@,$${priority},g"			|\
	$(SED) -e "s,@PHP_VERSION@,$(PHP_NAME_VERSION),g"	|\
	$(SED) -e "s,@PHP_MAJOR@,$(PHP_MAJOR_VERSION),g"	|\
	$(SED) -e "s,@PHP_MINOR@,$(PHP_MINOR_VERSION),g"	|\
	$(SED) -e "s,@PHP_RELEASE@,$(PHP_RELEASE_VERSION),g"	|\
	$(SED) -e "s,@PHP_API@,$(PHP_ZEND_VERSION),g"

SED_REPLACEMENT := $(SED) -e "/\#EXTRA\#/ r debian/$${versionless}.$${tmpl}.extra" | $(SED_VARIABLES) | $(SED) -e '/\#EXTRA\#/ d'

SAPI_PACKAGE := \
	case $${sapi} in \
	  embed)         package=$(PHP_LIBEMBED); source=libphp-$${sapi} ;; \
	  apache2)       package=$(PHP_APACHE2); source=libapache2-mod-php ;; \
	  *)             package=php$(PHP_NAME_VERSION)-$${sapi}; source=php-$${sapi} ;; \
	esac; \
	versionless=$$(echo $${package} | $(SED) -e 's/$(PHP_NAME_VERSION)//g');

MODULE_PACKAGE := \
	package=php$(PHP_NAME_VERSION)-$${module}; \
	versionless=$$(echo $${package} | $(SED) -e 's/$(PHP_NAME_VERSION)//g');

LIBTOOL_VERSION := $(shell dpkg-query -f'$${Version}' -W libtool)

# Disable the test now
RUN_TESTS := no
ifeq (nocheck,$(filter nocheck,$(DEB_BUILD_PROFILES)))
  $(warning Disabling tests due DEB_BUILD_PROFILES)
  DEB_BUILD_OPTIONS += nocheck
  RUN_TESTS := no
else
  ifeq (nocheck,$(filter nocheck,$(DEB_BUILD_OPTIONS)))
    $(warning Disabling tests due DEB_BUILD_OPTIONS)
    RUN_TESTS := no
  endif
endif
ifeq (,$(filter $(DEB_HOST_ARCH),$(SANE_ARCHS)))
  $(warning Disabling checks on $(DEB_HOST_ARCH))
  RUN_TESTS := no
endif

CONFIGURE_DTRACE_ARGS := --disable-dtrace

ifeq ($(DEB_HOST_ARCH_OS),linux)
  CONFIGURE_SYSTEMD := --with-fpm-systemd
  CONFIGURE_APPARMOR := --with-fpm-apparmor
endif

# specify some options to our patch system
QUILT_DIFF_OPTS := -p
QUILT_NO_DIFF_TIMESTAMPS := 1
export QUILT_DIFF_OPTS
export QUILT_NO_DIFF_TIMESTAMPS

export PROG_SENDMAIL := /usr/sbin/sendmail
ifeq (,$(findstring noopt,$(DEB_BUILD_OPTIONS)))
  DEB_CFLAGS_MAINT_APPEND += -O2
else
  DEB_CFLAGS_MAINT_APPEND += -O0
endif
DEB_CFLAGS_MAINT_APPEND += -Wall -pedantic -fsigned-char -fno-strict-aliasing
DEB_CFLAGS_MAINT_APPEND += $(shell getconf LFS_CFLAGS)

# OpenSSL 3.0 support
DEB_CFLAGS_MAINT_APPEND += -DOPENSSL_SUPPRESS_DEPRECATED

# Enable IEEE-conformant floating point math on alphas (not the default)
ifeq (alpha-linux-gnu,$(DEB_HOST_GNU_TYPE))
  DEB_CFLAGS_MAINT_APPEND += -mieee
endif

# Enable producing of debugging information
DEB_CFLAGS_MAINT_APPEND += -g

DEB_LDFLAGS_MAINT_APPEND = -Wl,--as-needed

export DEB_CFLAGS_MAINT_APPEND
export DEB_LDFLAGS_MAINT_APPEND

# some other helpful (for readability at least) shorthand variables
PHPIZE_BUILDDIR := /usr/lib/php/$(PHP_ZEND_VERSION)/build/

COMMON_CONFIG := \
		--build=$(DEB_BUILD_GNU_TYPE) \
		--host=$(DEB_HOST_GNU_TYPE) \
		--config-cache --cache-file=$(CURDIR)/config.cache \
		--libdir=\$${prefix}/lib/php \
		--libexecdir=\$${prefix}/lib/php \
		--datadir=\$${prefix}/share/php/$(PHP_NAME_VERSION) \
		--program-suffix=$(PHP_NAME_VERSION) \
		--sysconfdir=/etc \
		--localstatedir=/var \
		--mandir=/usr/share/man \
		--disable-all \
		--disable-debug \
		--disable-rpath \
		--disable-static \
		--with-pic \
		--with-layout=GNU \
		--without-pear \
		--enable-filter \
		--with-openssl=yes \
		--with-pcre-regex=/usr \
		--enable-hash \
		  --with-mhash=/usr \
		--enable-libxml \
		--enable-session \
		--with-system-tzdata \
		--with-zlib=/usr \
		  --with-zlib-dir=/usr \
		$(CONFIGURE_ZTS) \
		$(CONFIGURE_DTRACE_ARGS)

# disable all SAPIs in extension build
export ext_config = \
		--prefix=/usr --enable-cli --disable-cgi --disable-phpdbg \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/apache2 \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/apache2/conf.d \
		$(COMMON_CONFIG)

export apache2_config = \
		--prefix=/usr --with-apxs2=/usr/bin/apxs2 --enable-cli --disable-cgi --disable-phpdbg \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/apache2 \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/apache2/conf.d \
		$(COMMON_CONFIG)

export cgi_config = \
		--prefix=/usr --enable-cgi --enable-cli --disable-phpdbg \
		--enable-force-cgi-redirect --enable-fastcgi \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/cgi \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/cgi/conf.d \
		$(COMMON_CONFIG) \
		--enable-pcntl

export cli_config = \
		--prefix=/usr --enable-cli --disable-cgi --disable-phpdbg \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/cli \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/cli/conf.d \
		$(COMMON_CONFIG) \
		--enable-pcntl \
		--with-libedit=shared,/usr

export embed_config = \
		--prefix=/usr --enable-embed --enable-cli --disable-cgi --disable-phpdbg \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/embed \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/embed/conf.d \
		$(COMMON_CONFIG) \
		--without-mm \
		--enable-pcntl

export fpm_config = \
		--prefix=/usr --enable-fpm --enable-cli --disable-cgi --disable-phpdbg \
		--sysconfdir=/etc/php/$(PHP_NAME_VERSION)/fpm \
		--with-fpm-user=www-data --with-fpm-group=www-data \
		--with-fpm-acl \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/fpm \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/fpm/conf.d \
		$(COMMON_CONFIG) \
		--with-libevent-dir=/usr \
		$(CONFIGURE_SYSTEMD) \
		$(CONFIGURE_APPARMOR)

export phpdbg_config = \
		--prefix=/usr --enable-phpdbg --enable-cli --disable-cgi \
		--with-config-file-path=/etc/php/$(PHP_NAME_VERSION)/phpdbg \
		--with-config-file-scan-dir=/etc/php/$(PHP_NAME_VERSION)/phpdbg/conf.d \
		$(COMMON_CONFIG) \
		--enable-pcntl \
		--with-libedit=shared,/usr

BUILTIN_EXTENSION_CHECK=$$e=get_loaded_extensions(); natcasesort($$e); \
			$$s="The following extensions are built in:"; \
			foreach($$e as $$i) { $$s .= " $$i"; } \
			echo("php:Extensions=" . wordwrap($$s . ".\n", 75, "\$${Newline}"));
