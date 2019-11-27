#!/usr/bin/python -u

import argparse
import hashlib
import logging
import os
import re
import shutil
import subprocess
import sys
import tempfile
from datetime import datetime

################
#### Variables
################

# Packaging variables
PACKAGE_NAME = "notify4g"
INSTALL_ROOT_DIR = "/usr/bin"
LOG_DIR = "/var/log/notify4g"
SCRIPT_DIR = "/usr/lib/notify4g/scripts"
CONFIG_DIR = "/etc/notify4g"
CONFIG_FILE = "notify4g.toml"
CONFIG_DIR_D = "/etc/notify4g/notify4g.d"
LOGROTATE_DIR = "/etc/logrotate.d"
LOGROTATE_FILE = "notify4g"

RUN_USER = "rigaga"

INIT_SCRIPT = "scripts/init.sh"
SYSTEMD_SCRIPT = "scripts/notify4g.service"
LOGROTATE_SCRIPT = "etc/logrotate.d/notify4g"
DEFAULT_CONFIG = "etc/{}/" + CONFIG_FILE
DEFAULT_WINDOWS_CONFIG = "etc/notify4g_windows.conf"
POSTINST_SCRIPT = "scripts/post-install.sh"
PREINST_SCRIPT = "scripts/pre-install.sh"
POSTREMOVE_SCRIPT = "scripts/post-remove.sh"
PREREMOVE_SCRIPT = "scripts/pre-remove.sh"

# Default AWS S3 bucket for uploads
DEFAULT_BUCKET = ""

CONFIGURATION_FILES = [
    CONFIG_DIR + '/' + CONFIG_FILE,
    LOGROTATE_DIR + '/' + LOGROTATE_FILE,
]

# META-PACKAGE VARIABLES
PACKAGE_LICENSE = "MIT"
PACKAGE_URL = "https://github.com/gobars/rigaga"
MAINTAINER = "support@bjca.org.cn"
VENDOR = "RigagaData"
DESCRIPTION = "rigaga for notify4g"

# SCRIPT START
prereqs = ['git', 'go']
go_vet_command = "go tool vet -composites=true ./"
optional_prereqs = ['gvm', 'fpm', 'rpmbuild']

fpm_common_args = "-f -s dir --log error \
 --vendor {} \
 --url {} \
 --license {} \
 --maintainer {} \
 --config-files {} \
 --config-files {} \
 --after-install {} \
 --before-install {} \
 --after-remove {} \
 --before-remove {} \
 --description \"{}\"".format(
    VENDOR,
    PACKAGE_URL,
    PACKAGE_LICENSE,
    MAINTAINER,
    CONFIG_DIR + '/' + CONFIG_FILE,
    LOGROTATE_DIR + '/notify4g',
    POSTINST_SCRIPT,
    PREINST_SCRIPT,
    POSTREMOVE_SCRIPT,
    PREREMOVE_SCRIPT,
    DESCRIPTION)

targets = {
    'notify4g': './cmd/notify4g',
}

supported_builds = {
    "windows": ["amd64", "i386"],
    "linux": ["amd64", "i386", "armhf", "armel", "arm64", "static_amd64", "s390x"],
    "freebsd": ["amd64", "i386"]
}

supported_packages = {
    "linux": ["rpm"],
    "windows": ["zip"],
    "freebsd": ["tar"]
}

next_version = '1.0.0'


################
#### notify4g Functions
################

def print_banner():
    logging.info("""

__________ .__                                   
\______   \|__|   ____  _____      ____  _____   
 |       _/|  |  / ___\ \__  \    / ___\ \__  \  
 |    |   \|  | / /_/  > / __ \_ / /_/  > / __ \_
 |____|_  /|__| \___  / (____  / \___  / (____  /
        \/     /_____/       \/ /_____/       \/ 

 Build Script
""")


def create_package_fs(build_root):
    """Create a filesystem structure to mimic the package filesystem.
    """
    logging.debug("Creating a filesystem hierarchy from directory: {}".format(build_root))
    # Using [1:] for the path names due to them being absolute
    # (will overwrite previous paths, per 'os.path.join' documentation)
    dirs = [INSTALL_ROOT_DIR[1:], LOG_DIR[1:], SCRIPT_DIR[1:], CONFIG_DIR[1:], LOGROTATE_DIR[1:], CONFIG_DIR_D[1:]]
    for d in dirs:
        os.makedirs(os.path.join(build_root, d))
        os.chmod(os.path.join(build_root, d), 0o755)


def package_scripts(build_root, role, config_only=False, windows=False):
    """Copy the necessary scripts and configuration files to the package
    filesystem.
    """
    logging.info("build role for {}".format(role))
    if config_only or windows:
        logging.info("Copying configuration to build directory")
        if windows:
            shutil.copyfile(DEFAULT_WINDOWS_CONFIG, os.path.join(build_root, CONFIG_FILE))
        else:
            shutil.copyfile(DEFAULT_CONFIG, os.path.join(build_root, CONFIG_FILE))
        os.chmod(os.path.join(build_root, CONFIG_FILE), 0o644)
    else:
        logging.info("Copying scripts and configuration to build directory")
        shutil.copyfile(INIT_SCRIPT, os.path.join(build_root, SCRIPT_DIR[1:], INIT_SCRIPT.split('/')[1]))
        os.chmod(os.path.join(build_root, SCRIPT_DIR[1:], INIT_SCRIPT.split('/')[1]), 0o644)
        shutil.copyfile(SYSTEMD_SCRIPT, os.path.join(build_root, SCRIPT_DIR[1:], SYSTEMD_SCRIPT.split('/')[1]))
        os.chmod(os.path.join(build_root, SCRIPT_DIR[1:], SYSTEMD_SCRIPT.split('/')[1]), 0o644)
        shutil.copyfile(LOGROTATE_SCRIPT, os.path.join(build_root, LOGROTATE_DIR[1:], LOGROTATE_FILE))
        os.chmod(os.path.join(build_root, LOGROTATE_DIR[1:], LOGROTATE_FILE), 0o644)
        shutil.copyfile(DEFAULT_CONFIG.format(role), os.path.join(build_root, CONFIG_DIR[1:], CONFIG_FILE))
        os.chmod(os.path.join(build_root, CONFIG_DIR[1:], CONFIG_FILE), 0o644)


def run_generate():
    # NOOP for rigaga
    return True


def go_get(branch, update=False, no_uncommitted=False):
    """Retrieve build dependencies or restore pinned dependencies.
    """
    if local_changes() and no_uncommitted:
        logging.error("There are uncommitted changes in the current directory.")
        return False
    return True


def run(command, allow_failure=False, shell=False):
    """Run shell command (convenience wrapper around subprocess).
    """
    out = None
    logging.debug("{}".format(command))
    try:
        if shell:
            out = subprocess.check_output(command, stderr=subprocess.STDOUT, shell=shell)
        else:
            out = subprocess.check_output(command.split(), stderr=subprocess.STDOUT)
        out = out.decode('utf-8').strip()
        # logging.debug("Command output: {}".format(out))
    except subprocess.CalledProcessError as e:
        if allow_failure:
            logging.warn("Command '{}' failed with error: {}".format(command, e.output))
            return None
        else:
            logging.error("Command '{}' failed with error: {}".format(command, e.output))
            sys.exit(1)
    except OSError as e:
        if allow_failure:
            logging.warn("Command '{}' failed with error: {}".format(command, e))
            return out
        else:
            logging.error("Command '{}' failed with error: {}".format(command, e))
            sys.exit(1)
    else:
        return out


def create_temp_dir(prefix=None):
    """ Create temporary directory with optional prefix.
    """
    if prefix is None:
        return tempfile.mkdtemp(prefix="{}-build.".format(PACKAGE_NAME))
    else:
        return tempfile.mkdtemp(prefix=prefix)


def get_system_arch():
    """Retrieve current system architecture.
    """
    arch = os.uname()[4]
    if arch == "x86_64":
        arch = "amd64"
    elif arch == "386":
        arch = "i386"
    elif "arm64" in arch:
        arch = "arm64"
    elif 'arm' in arch:
        # Prevent uname from reporting full ARM arch (eg 'armv7l')
        arch = "arm"
    return arch


def get_system_platform():
    """Retrieve current system platform.
    """
    if sys.platform.startswith("linux"):
        return "linux"
    else:
        return sys.platform


def get_current_branch():
    """Retrieve the current git branch.
    """
    command = "git rev-parse --abbrev-ref HEAD"
    out = run(command)
    return out.strip()


def local_changes():
    """Return True if there are local un-committed changes.
    """
    output = run("git diff-files --ignore-submodules --").strip()
    if len(output) > 0:
        return True
    return False


def get_current_commit(short=False):
    """Retrieve the current git commit.
    """
    command = None
    if short:
        command = "git log --pretty=format:'%h' -n 1"
    else:
        command = "git rev-parse HEAD"
    out = run(command)
    return out.strip('\'\n\r ')


def get_current_version():
    """Parse version information from git tag output.
    """
    version_tag = get_current_version_tag()
    if not version_tag:
        return None
    # Remove leading 'v'
    if version_tag[0] == 'v':
        version_tag = version_tag[1:]
    # Replace any '-'/'_' with '~'
    if '-' in version_tag:
        version_tag = version_tag.replace("-", "~")
    if '_' in version_tag:
        version_tag = version_tag.replace("_", "~")
    return version_tag


def get_current_version_tag():
    """Retrieve the raw git version tag.
    """
    version = run("git describe --exact-match --tags 2>/dev/null",
                  allow_failure=True, shell=True)
    return version


def get_go_version():
    """Retrieve version information for Go.
    """
    out = run("go version")
    matches = re.search('go version go(\S+)', out)
    if matches is not None:
        return matches.groups()[0].strip()
    return None


def check_path_for(b):
    """Check the the user's path for the provided binary.
    """

    def is_exe(fpath):
        return os.path.isfile(fpath) and os.access(fpath, os.X_OK)

    for path in os.environ["PATH"].split(os.pathsep):
        path = path.strip('"')
        full_path = os.path.join(path, b)
        if os.path.isfile(full_path) and os.access(full_path, os.X_OK):
            return full_path


def check_environ(build_dir=None):
    """Check environment for common Go variables.
    """
    logging.info("Checking environment...")
    for v in ["GOPATH", "GOBIN", "GOROOT"]:
        logging.debug("Using '{}' for {}".format(os.environ.get(v), v))

    cwd = os.getcwd()
    if build_dir is None and os.environ.get("GOPATH") and os.environ.get("GOPATH") not in cwd:
        logging.warn("Your current directory is not under your GOPATH. This may lead to build failures.")
    return True


def check_prereqs():
    """Check user path for required dependencies.
    """
    logging.info("Checking for dependencies...")
    for req in prereqs:
        if not check_path_for(req):
            logging.error("Could not find dependency: {}".format(req))
            return False
    return True


def build(version=None,
          platform=None,
          arch=None,
          nightly=False,
          race=False,
          clean=False,
          outdir=".",
          tags=[],
          static=False):
    """Build each target for the specified architecture and platform.
    """
    logging.info("Starting build for {}/{}...".format(platform, arch))
    logging.info("Using Go version: {}".format(get_go_version()))
    logging.info("Using git branch: {}".format(get_current_branch()))
    logging.info("Using git commit: {}".format(get_current_commit()))
    if static:
        logging.info("Using statically-compiled output.")
    if race:
        logging.info("Race is enabled.")
    if len(tags) > 0:
        logging.info("Using build tags: {}".format(','.join(tags)))

    logging.info("Sending build output to: {}".format(outdir))
    if not os.path.exists(outdir):
        os.makedirs(outdir)
    elif clean and outdir != '/' and outdir != ".":
        logging.info("Cleaning build directory '{}' before building.".format(outdir))
        shutil.rmtree(outdir)
        os.makedirs(outdir)

    logging.info("Using version '{}' for build.".format(version))

    tmp_build_dir = create_temp_dir()
    for target, path in targets.items():
        logging.info("Building target: {}".format(target))
        build_command = ""

        # Handle static binary output
        if static is True or "static_" in arch:
            if "static_" in arch:
                static = True
                arch = arch.replace("static_", "")
            build_command += "CGO_ENABLED=0 "

        # Handle variations in architecture output
        if arch == "i386" or arch == "i686":
            arch = "386"
        elif "arm64" in arch:
            arch = "arm64"
        elif "arm" in arch:
            arch = "arm"
        build_command += "GOOS={} GOARCH={} ".format(platform, arch)

        if "arm" in arch:
            if arch == "armel":
                build_command += "GOARM=5 "
            elif arch == "armhf" or arch == "arm":
                build_command += "GOARM=6 "
            elif arch == "arm64":
                # TODO(rossmcdonald) - Verify this is the correct setting for arm64
                build_command += "GOARM=7 "
            else:
                logging.error("Invalid ARM architecture specified: {}".format(arch))
                logging.error("Please specify either 'armel', 'armhf', or 'arm64'.")
                return False
        if platform == 'windows':
            target = target + '.exe'
        build_command += "go build -o {} ".format(os.path.join(outdir, target))
        if race:
            build_command += "-race "
        if len(tags) > 0:
            build_command += "-tags {} ".format(','.join(tags))

        ldflags = [
            '-w', '-s',
            '-X', 'main.branch={}'.format(get_current_branch()),
            '-X', 'main.commit={}'.format(get_current_commit(short=True))]
        if version:
            ldflags.append('-X')
            ldflags.append('main.version={}'.format(version))
        build_command += ' -ldflags="{}" '.format(' '.join(ldflags))

        if static:
            build_command += " -a -installsuffix cgo "
        build_command += path
        start_time = datetime.utcnow()
        run(build_command, shell=True)
        end_time = datetime.utcnow()
        logging.info("Time taken: {}s".format((end_time - start_time).total_seconds()))
    return True


def generate_sha256_from_file(path):
    """Generate SHA256 hash signature based on the contents of the file at path.
    """
    m = hashlib.sha256()
    with open(path, 'rb') as f:
        m.update(f.read())
    return m.hexdigest()


def package(build_output, pkg_name, version, nightly=False, iteration=1, static=False, release=False, role="agent"):
    """Package the output of the build process.
    """
    outfiles = []
    tmp_build_dir = create_temp_dir()
    logging.debug("Packaging for build output: {}".format(build_output))
    logging.info("Using temporary directory: {}".format(tmp_build_dir))
    try:
        for platform in build_output:
            # Create top-level folder displaying which platform (linux, etc)
            os.makedirs(os.path.join(tmp_build_dir, platform))
            for arch in build_output[platform]:
                logging.info("Creating packages for {}/{}".format(platform, arch))
                # Create second-level directory displaying the architecture (amd64, etc)
                current_location = build_output[platform][arch]

                # Create directory tree to mimic file system of package
                build_root = os.path.join(tmp_build_dir,
                                          platform,
                                          arch,
                                          PACKAGE_NAME)
                os.makedirs(build_root)

                # Copy packaging scripts to build directory
                if platform == "windows":
                    # For windows and static builds, just copy
                    # binaries to root of package (no other scripts or
                    # directories)
                    package_scripts(build_root, config_only=True, windows=True)
                elif static or "static_" in arch:
                    package_scripts(build_root, role, config_only=True)
                else:
                    create_package_fs(build_root)
                    package_scripts(build_root, role)

                for binary in targets:
                    # Copy newly-built binaries to packaging directory
                    if platform == 'windows':
                        binary = binary + '.exe'
                    if platform == 'windows' or static or "static_" in arch:
                        # Where the binary should go in the package filesystem
                        to = os.path.join(build_root, binary)
                        # Where the binary currently is located
                        fr = os.path.join(current_location, binary)
                    else:
                        # Where the binary currently is located
                        fr = os.path.join(current_location, binary)
                        # Where the binary should go in the package filesystem
                        to = os.path.join(build_root, INSTALL_ROOT_DIR[1:], binary)
                    shutil.copy(fr, to)

                for package_type in supported_packages[platform]:
                    # Package the directory structure for each package type for the platform
                    logging.debug("Packaging directory '{}' as '{}'.".format(build_root, package_type))
                    name = pkg_name
                    # Reset version, iteration, and current location on each run
                    # since they may be modified below.
                    package_version = version
                    package_iteration = iteration
                    if "static_" in arch:
                        # Remove the "static_" from the displayed arch on the package
                        package_arch = arch.replace("static_", "")
                    elif package_type == "rpm" and arch == 'armhf':
                        package_arch = 'armv6hl'
                    else:
                        package_arch = arch
                    if not version:
                        package_version = "{}".format(next_version)
                        package_iteration = "0"
                    package_build_root = build_root
                    current_location = build_output[platform][arch]

                    if package_type in ['zip', 'tar']:
                        # For tars and zips, start the packaging one folder above
                        # the build root (to include the package name)
                        package_build_root = os.path.join('/', '/'.join(build_root.split('/')[:-1]))
                        if nightly:
                            if static or "static_" in arch:
                                name = '{}-static-nightly_{}_{}'.format(name,
                                                                        platform,
                                                                        package_arch)
                            else:
                                name = '{}-nightly_{}_{}'.format(name,
                                                                 platform,
                                                                 package_arch)
                        else:
                            if static or "static_" in arch:
                                name = '{}-{}-static_{}_{}'.format(name,
                                                                   package_version,
                                                                   platform,
                                                                   package_arch)
                            else:
                                name = '{}-{}_{}_{}'.format(name,
                                                            package_version,
                                                            platform,
                                                            package_arch)
                        current_location = os.path.join(os.getcwd(), current_location)
                        if package_type == 'tar':
                            tar_command = "cd {} && tar -cvzf {}.tar.gz ./*".format(package_build_root, name)
                            run(tar_command, shell=True)
                            run("mv {}.tar.gz {}".format(os.path.join(package_build_root, name), current_location),
                                shell=True)
                            outfile = os.path.join(current_location, name + ".tar.gz")
                            outfiles.append(outfile)
                        elif package_type == 'zip':
                            zip_command = "cd {} && zip -r {}.zip ./*".format(package_build_root, name)
                            run(zip_command, shell=True)
                            run("mv {}.zip {}".format(os.path.join(package_build_root, name), current_location),
                                shell=True)
                            outfile = os.path.join(current_location, name + ".zip")
                            outfiles.append(outfile)
                    elif package_type not in ['zip', 'tar'] and static or "static_" in arch:
                        logging.info("Skipping package type '{}' for static builds.".format(package_type))
                    else:
                        if package_type == 'rpm' and release and '~' in package_version:
                            package_version, suffix = package_version.split('~', 1)
                            # The ~ indicatees that this is a prerelease so we give it a leading 0.
                            package_iteration = "0.%s" % suffix
                        fpm_command = "fpm {} --name {} -a {} -t {} --rpm-os linux --version {} --iteration {} -C {} -p {} ".format(
                            fpm_common_args,
                            name,
                            package_arch,
                            package_type,
                            package_version,
                            role,
                            package_build_root,
                            current_location)
                        if package_type == "rpm":
                            fpm_command += "--depends coreutils --depends shadow-utils --rpm-posttrans {}".format(
                                POSTINST_SCRIPT)
                        logging.info("fpm_command: {}".format(fpm_command))

                        out = run(fpm_command, shell=True)
                        matches = re.search(':path=>"(.*)"', out)
                        outfile = None
                        if matches is not None:
                            outfile = matches.groups()[0]
                        if outfile is None:
                            logging.warn("Could not determine output from packaging output!")
                        else:
                            if nightly:
                                # Strip nightly version from package name
                                new_outfile = outfile.replace("{}-{}".format(package_version, package_iteration),
                                                              "nightly")
                                os.rename(outfile, new_outfile)
                                outfile = new_outfile
                            else:
                                if package_type == 'rpm':
                                    # rpm's convert any dashes to underscores
                                    package_version = package_version.replace("-", "_")
                            outfiles.append(os.path.join(os.getcwd(), outfile))
        logging.debug("Produced package files: {}".format(outfiles))
        return outfiles
    finally:
        # Cleanup
        shutil.rmtree(tmp_build_dir)


def main(args):
    global PACKAGE_NAME

    if args.release and args.nightly:
        logging.error("Cannot be both a nightly and a release.")
        return 1

    if args.nightly:
        args.iteration = 0

    # Pre-build checks
    check_environ()
    if not check_prereqs():
        return 1
    if args.build_tags is None:
        args.build_tags = []
    else:
        args.build_tags = args.build_tags.split(',')

    orig_commit = get_current_commit(short=True)
    orig_branch = get_current_branch()

    if args.platform not in supported_builds and args.platform != 'all':
        logging.error("Invalid build platform: {}".format(args.platform))
        return 1

    build_output = {}

    if args.branch != orig_branch and args.commit != orig_commit:
        logging.error("Can only specify one branch or commit to build from.")
        return 1
    elif args.branch != orig_branch:
        logging.info("Moving to git branch: {}".format(args.branch))
        run("git checkout {}".format(args.branch))
    elif args.commit != orig_commit:
        logging.info("Moving to git commit: {}".format(args.commit))
        run("git checkout {}".format(args.commit))

    if not args.no_get:
        if not go_get(args.branch, update=args.update, no_uncommitted=args.no_uncommitted):
            return 1

    if args.generate:
        if not run_generate():
            return 1

    if args.test:
        if not run_tests(args.race, args.parallel, args.timeout, args.no_vet):
            return 1

    platforms = []
    single_build = True
    if args.platform == 'all':
        platforms = supported_builds.keys()
        single_build = False
    else:
        platforms = [args.platform]

    for platform in platforms:
        build_output.update({platform: {}})
        archs = []
        if args.arch == "all":
            single_build = False
            archs = supported_builds.get(platform)
        else:
            archs = [args.arch]

        for arch in archs:
            od = args.outdir
            if not single_build:
                od = os.path.join(args.outdir, platform, arch)
            if not build(version=args.version,
                         platform=platform,
                         arch=arch,
                         nightly=args.nightly,
                         race=args.race,
                         clean=args.clean,
                         outdir=od,
                         tags=args.build_tags,
                         static=args.static):
                return 1
            build_output.get(platform).update({arch: od})

    # Build packages
    if args.package:
        if not check_path_for("fpm"):
            logging.error("FPM ruby gem required for packaging. Stopping.")
            return 1

        packages = package(build_output,
                           args.name,
                           args.version,
                           nightly=args.nightly,
                           iteration=args.iteration,
                           static=args.static,
                           release=args.release,
                           role=args.role)
        if args.sign:
            logging.debug("Generating GPG signatures for packages: {}".format(packages))
            sigs = []  # retain signatures so they can be uploaded with packages
            for p in packages:
                if generate_sig_from_file(p):
                    sigs.append(p + '.asc')
                else:
                    logging.error("Creation of signature for package [{}] failed!".format(p))
                    return 1
            packages += sigs
        if args.upload:
            logging.debug("Files staged for upload: {}".format(packages))
            if args.nightly:
                args.upload_overwrite = True
            if not upload_packages(packages, bucket_name=args.bucket, overwrite=args.upload_overwrite):
                return 1
        logging.info("Packages created:")
        for filename in packages:
            logging.info("%s (SHA256=%s)",
                         os.path.basename(filename),
                         generate_sha256_from_file(filename))
    if orig_branch != get_current_branch():
        logging.info("Moving back to original git branch: {}".format(args.branch))
        run("git checkout {}".format(orig_branch))

    return 0


if __name__ == '__main__':
    LOG_LEVEL = logging.INFO
    if '--debug' in sys.argv[1:]:
        LOG_LEVEL = logging.DEBUG
    log_format = '[%(levelname)s] %(funcName)s: %(message)s'
    logging.basicConfig(level=LOG_LEVEL,
                        format=log_format)

    parser = argparse.ArgumentParser(description='Notify4g build and packaging script.')
    parser.add_argument('--verbose', '-v', '--debug',
                        action='store_true',
                        help='Use debug output')
    parser.add_argument('--outdir', '-o',
                        metavar='<output directory>',
                        default='./build/',
                        type=os.path.abspath,
                        help='Output directory')
    parser.add_argument('--name', '-n',
                        metavar='<name>',
                        default=PACKAGE_NAME,
                        type=str,
                        help='Name to use for package name (when package is specified)')
    parser.add_argument('--arch',
                        metavar='<amd64|i386|armhf|arm64|armel|all>',
                        type=str,
                        default=get_system_arch(),
                        help='Target architecture for build output')
    parser.add_argument('--platform',
                        metavar='<linux|darwin|windows|all>',
                        type=str,
                        default=get_system_platform(),
                        help='Target platform for build output')
    parser.add_argument('--branch',
                        metavar='<branch>',
                        type=str,
                        default=get_current_branch(),
                        help='Build from a specific branch')
    parser.add_argument('--commit',
                        metavar='<commit>',
                        type=str,
                        default=get_current_commit(short=True),
                        help='Build from a specific commit')
    parser.add_argument('--version',
                        metavar='<version>',
                        type=str,
                        default=get_current_version(),
                        help='Version information to apply to build output (ex: 0.12.0)')
    parser.add_argument('--iteration',
                        metavar='<package iteration>',
                        type=str,
                        default="1",
                        help='Package iteration to apply to build output (defaults to 1)')
    parser.add_argument('--stats',
                        action='store_true',
                        help='Emit build metrics (requires Notify4g Python client)')
    parser.add_argument('--stats-server',
                        metavar='<hostname:port>',
                        type=str,
                        help='Send build stats to Notify4g using provided hostname and port')
    parser.add_argument('--stats-db',
                        metavar='<database name>',
                        type=str,
                        help='Send build stats to Notify4g using provided database name')
    parser.add_argument('--nightly',
                        action='store_true',
                        help='Mark build output as nightly build (will incremement the minor version)')
    parser.add_argument('--update',
                        action='store_true',
                        help='Update build dependencies prior to building')
    parser.add_argument('--package',
                        action='store_true',
                        help='Package binary output')
    parser.add_argument('--release',
                        action='store_true',
                        help='Mark build output as release')
    parser.add_argument('--clean',
                        action='store_true',
                        help='Clean output directory before building')
    parser.add_argument('--no-get',
                        action='store_true',
                        help='Do not retrieve pinned dependencies when building')
    parser.add_argument('--no-uncommitted',
                        action='store_true',
                        help='Fail if uncommitted changes exist in the working directory')
    parser.add_argument('--upload',
                        action='store_true',
                        help='Upload output packages to AWS S3')
    parser.add_argument('--upload-overwrite', '-w',
                        action='store_true',
                        help='Upload output packages to AWS S3')
    parser.add_argument('--bucket',
                        metavar='<S3 bucket name>',
                        type=str,
                        default=DEFAULT_BUCKET,
                        help='Destination bucket for uploads')
    parser.add_argument('--generate',
                        action='store_true',
                        help='Run "go generate" before building')
    parser.add_argument('--build-tags',
                        metavar='<tags>',
                        help='Optional build tags to use for compilation')
    parser.add_argument('--static',
                        action='store_true',
                        help='Create statically-compiled binary output')
    parser.add_argument('--sign',
                        action='store_true',
                        help='Create GPG detached signatures for packages (when package is specified)')
    parser.add_argument('--test',
                        action='store_true',
                        help='Run tests (does not produce build output)')
    parser.add_argument('--no-vet',
                        action='store_true',
                        help='Do not run "go vet" when running tests')
    parser.add_argument('--race',
                        action='store_true',
                        help='Enable race flag for build output')
    parser.add_argument('--parallel',
                        metavar='<num threads>',
                        type=int,
                        help='Number of tests to run simultaneously')
    parser.add_argument('--timeout',
                        metavar='<timeout>',
                        type=str,
                        help='Timeout for tests before failing')
    parser.add_argument('--role',
                        metavar='<agent|transform>',
                        default='agent',
                        type=str,
                        help='Package Role')

    args = parser.parse_args()
    print_banner()
    sys.exit(main(args))
