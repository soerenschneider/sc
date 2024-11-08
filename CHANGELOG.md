# Changelog

## [1.2.1](https://github.com/soerenschneider/sc/compare/v1.2.0...v1.2.1) (2024-11-08)


### Bug Fixes

* **deps:** bump github.com/soerenschneider/sc-agent from 1.4.0 to 1.8.0 ([2dc18c0](https://github.com/soerenschneider/sc/commit/2dc18c086b2bea1eaa83ff2b658e981647df7fbd))

## [1.2.0](https://github.com/soerenschneider/sc/compare/v1.1.1...v1.2.0) (2024-11-08)


### Features

* **sc-agent:** automatically set default protocol and port for server if not supplied ([832ef50](https://github.com/soerenschneider/sc/commit/832ef5060edea755ff5822112c631a17bff8b2b3))
* **sc-agent:** expand files to support paths using a tilde prefix ([aee1039](https://github.com/soerenschneider/sc/commit/aee10396ae972998f54928f8dbb72d9edb2ca5de))


### Bug Fixes

* **deps:** bump github.com/jedib0t/go-pretty/v6 from 6.6.0 to 6.6.1 ([31e95fb](https://github.com/soerenschneider/sc/commit/31e95fb15c9aeedc48a9099f27afb99692cf6053))

## [1.1.1](https://github.com/soerenschneider/sc/compare/v1.1.0...v1.1.1) (2024-10-11)


### Bug Fixes

* adapt to change ([76ad84c](https://github.com/soerenschneider/sc/commit/76ad84c00c665f4fdb09c38c89cd63ffbd0b8f02))
* **agent:** bump github.com/hashicorp/vault/api from 1.14.0 to 1.15.0 ([8e5d574](https://github.com/soerenschneider/sc/commit/8e5d5746665a00b89b568344e12d1d6e4d2f2046))
* **agent:** bump github.com/jedib0t/go-pretty/v6 from 6.5.9 to 6.6.0 ([fe6c780](https://github.com/soerenschneider/sc/commit/fe6c7803d183eab049f60e30567efdd3eaed6b8d))
* **agent:** bump github.com/soerenschneider/sc-agent from 1.0.1 to 1.4.0 ([346cc25](https://github.com/soerenschneider/sc/commit/346cc252f02fa5b17a2759980c2bb557e1841be6))
* **agent:** bump github.com/soerenschneider/vault-pki-cli ([71e0259](https://github.com/soerenschneider/sc/commit/71e025920c45b56c80963f56b66490c7d012a8b8))
* **agent:** bump github.com/soerenschneider/vault-ssh-cli ([6cf0914](https://github.com/soerenschneider/sc/commit/6cf0914d7e95c4c51a128cdbb7e0f066b2e661f9))
* **agent:** bump golang from 1.23.0 to 1.23.1 ([fd2f173](https://github.com/soerenschneider/sc/commit/fd2f1736c741941208f6db909d0c213e9021048f))
* **agent:** bump golang from 1.23.1 to 1.23.2 ([c49c32e](https://github.com/soerenschneider/sc/commit/c49c32efda644dd25fa47dfe69896cfa6516b7ae))
* **agent:** bump golang.org/x/term from 0.23.0 to 0.24.0 ([5bd6ed8](https://github.com/soerenschneider/sc/commit/5bd6ed838ccbc991f2591f99d11a8e1cdb6e653b))
* **agent:** bump golang.org/x/term from 0.24.0 to 0.25.0 ([c58afca](https://github.com/soerenschneider/sc/commit/c58afca5c38dbd9b571c5b9737f6e78dfb4e7d7b))
* don't require passing a ca cert ([dbb4211](https://github.com/soerenschneider/sc/commit/dbb42116ab03540b670ca794a5a1fafc85b1015e))
* **lint:** check err ([2a92323](https://github.com/soerenschneider/sc/commit/2a9232397d77c3058bb536420c2a0c72b832b712))

## [1.1.0](https://github.com/soerenschneider/sc/compare/v1.0.0...v1.1.0) (2024-09-04)


### Features

* better logging ([909b4b4](https://github.com/soerenschneider/sc/commit/909b4b4235a226953cb388bc3a7300cafc5f518b))


### Bug Fixes

* bump sc-agent to v1.0.1 ([01f1109](https://github.com/soerenschneider/sc/commit/01f11090802f7ac8a2829ee36b7808c196387093))

## 1.0.0 (2024-09-02)


### Features

* add 'version' subcommand ([3554c3b](https://github.com/soerenschneider/sc/commit/3554c3b11ce64e66ff821cd53a2ff5ba6abdd427))
* check for updated releases on start ([f256c2f](https://github.com/soerenschneider/sc/commit/f256c2f44ec675610a8316681ca7db15fef1787c))
* initial support for 'pki' subcommand ([4759e3a](https://github.com/soerenschneider/sc/commit/4759e3a1fa82d1d5cb17fce9e0b0845f287c81ac))


### Bug Fixes

* set min tls version to 1.3 ([11edcfe](https://github.com/soerenschneider/sc/commit/11edcfe616dc69a58474d08481601124aae5f676))
