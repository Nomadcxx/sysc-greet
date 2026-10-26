# Native FreeBSD prerequisite

Build the native greetd prerequisite for the FreeBSD/GhostBSD port. The graphical sysc-greet port is still under development; its current metrics dependency requires Linux.

On FreeBSD, install Rust and run:

```sh
sudo pkg install rust
./scripts/build-greetd-freebsd.sh dist/freebsd-greetd
```

The script verifies the source archives, applies the patches, checks the PAM binding against the installed system headers, runs the upstream tests and builds release binaries. It writes `greetd` and `agreety` to the output directory. You can use `CARGO_TARGET_DIR` for a build cache.

The greetd patch builds on [Daniel Walter's FreeBSD patch](https://lists.sr.ht/~kennylevinsen/greetd-devel/patches/58445/mbox), adapted to upstream commit `d5f27c3a8b1758da9e45608d2cff20303f21c2ff`. It covers native console numbering, executable handoff, runtime paths, PAM configuration lookup, native login classes and process cleanup. The separate pam-sys patch supplies OpenPAM constants and error conversions; Linux-PAM values differ.

Native tests on FreeBSD 15.1-RELEASE-p6 passed incorrect-password rejection, unprivileged login, PAM runtime-directory ownership and permissions, VT switching, logout/greeter return and parent-death signalling. Native login-context checks also cover resource limits, class environment, canonical identity and supplementary groups added after initial group setup.

From a separately prepared patched greetd source checkout, these tests require root and console access. The build helper removes its temporary source tree:

```sh
cargo test -p greetd native_console_queries -- --ignored --test-threads=1
cargo test -p greetd native_login_context -- --ignored --test-threads=1
```

Run console and login validation in a VM or on a spare console with SSH access. FreeBSD uses `/dev/ttyvN`, `/var/run` and `/usr/local/etc/pam.d`; its base PAM `system` policy includes `pam_xdg`. These tests used a dedicated greeter account and the native PAM policy. GhostBSD still needs separate validation.
