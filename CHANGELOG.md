# Changelog

## [1.6.0](https://github.com/soerenschneider/sc/compare/v1.5.0...v1.6.0) (2025-07-08)


### Features

* read profile from env variable ([26e6e1f](https://github.com/soerenschneider/sc/commit/26e6e1f95a721ecfb539050ea5ffdf0bd92c62ba))
* unblock subcommands by running update check concurrently ([92d87e4](https://github.com/soerenschneider/sc/commit/92d87e4b50461b629af83f1925b3a79ea12adfe7))


### Bug Fixes

* fix condition ([0d7ef53](https://github.com/soerenschneider/sc/commit/0d7ef538cf2b95495dc41e909dbd396a494e82d2))

## [1.5.0](https://github.com/soerenschneider/sc/compare/v1.4.0...v1.5.0) (2025-07-02)


### Features

* add command hierarchy for profiles ([c7ec117](https://github.com/soerenschneider/sc/commit/c7ec11745c33ae69edf4d945e77c3297ea728482))
* add support for oidc ([f8c810f](https://github.com/soerenschneider/sc/commit/f8c810f7a15c45304d262d44aa0640e923febf58))
* add support torrents ([bf307d7](https://github.com/soerenschneider/sc/commit/bf307d75f0a6638dfdd111184c1eb9582ea58990))
* add userdata that remembers last used values for commands ([d5274e3](https://github.com/soerenschneider/sc/commit/d5274e3b4ebf6f4f40da0fe023d328ab0ccb020d))
* improve output ([2a78343](https://github.com/soerenschneider/sc/commit/2a78343d3f9d1340a256f34e2e9d1d3422cba816))
* support for linkding api ([1301d64](https://github.com/soerenschneider/sc/commit/1301d643b481de4be2e4532263cdad2b6f531f86))


### Bug Fixes

* actually only delete selected torrents ([e5d4e4f](https://github.com/soerenschneider/sc/commit/e5d4e4f44a706b88be17a137f1033b1cf9750851))
* **deps:** bump github.com/go-viper/mapstructure/v2 from 2.2.1 to 2.3.0 ([b31db04](https://github.com/soerenschneider/sc/commit/b31db04cdf6df337828425314cf68238e98726ee))
* **deps:** bump github.com/go-viper/mapstructure/v2 from 2.2.1 to 2.3.0 ([21ac91b](https://github.com/soerenschneider/sc/commit/21ac91b4d8d2323d02c47745d18ae7a10583807b))
* **deps:** bump github.com/hashicorp/go-retryablehttp ([b549d2e](https://github.com/soerenschneider/sc/commit/b549d2ea56a13c158a47e6f303137514a62b6ce2))
* **deps:** bump github.com/hashicorp/go-retryablehttp from 0.7.7 to 0.7.8 ([abc9165](https://github.com/soerenschneider/sc/commit/abc9165c28fe1af97f70ca66eece78bd9726fc22))
* **deps:** bump golang from 1.24.3 to 1.24.4 ([d39df20](https://github.com/soerenschneider/sc/commit/d39df201a25af783fa5a951ebc50654647decbc5))
* **deps:** bump golang from 1.24.3 to 1.24.4 ([9379c99](https://github.com/soerenschneider/sc/commit/9379c99e98a70ed87f6140ffd04a3cdace6dbb40))
* **deps:** bump golang.org/x/crypto from 0.38.0 to 0.39.0 ([cb1ca8a](https://github.com/soerenschneider/sc/commit/cb1ca8ab82749c903289c8144fbab7c5536e0e26))
* **deps:** bump golang.org/x/crypto from 0.38.0 to 0.39.0 ([d2b3223](https://github.com/soerenschneider/sc/commit/d2b3223644c3180fac51af9ac8924ef1db77b79e))
* fix build on linux(?) ([f5f49c6](https://github.com/soerenschneider/sc/commit/f5f49c698d9c82db8f8ec2ddb0efb19485f07ce2))
* improve ux by marking persistent flag 'addr' required ([b64ba75](https://github.com/soerenschneider/sc/commit/b64ba7590eb63d132ea7c3bf6dd1ecf794391e3a))
* resolve golangci-lint issues ([ab61070](https://github.com/soerenschneider/sc/commit/ab61070add9217b27d5e94ae08ae3c86d3733603))

## [1.4.0](https://github.com/soerenschneider/sc/compare/v1.3.0...v1.4.0) (2025-05-26)


### Features

* add subcommand to query victorialogs ([ba75997](https://github.com/soerenschneider/sc/commit/ba75997a12ca459a7061fb053c254fa361446007))
* add support for configuring healthchecks ([5e292d4](https://github.com/soerenschneider/sc/commit/5e292d40666a0c59aa083b52d67a184b26f796a9))
* add support for victorialogs ([509c83b](https://github.com/soerenschneider/sc/commit/509c83b5f9b663f649e9a1a073fabe575d016750))
* set max table size ([b7d5120](https://github.com/soerenschneider/sc/commit/b7d5120d125c591d45884313ce1de8e8f4a78322))
* support for profiles ([b308eb8](https://github.com/soerenschneider/sc/commit/b308eb811d5878a50804c60ed82fa879c5a62b29))


### Bug Fixes

* bump all dependencies ([ff52cf5](https://github.com/soerenschneider/sc/commit/ff52cf5901e37b8252eaacd852f5f770b22bc57f))
* do not exit prematurely if profile is supplied ([883ef9a](https://github.com/soerenschneider/sc/commit/883ef9afbb0864d6129a610192a8644543722ff4))
* print warning when profile is not found ([afe3df4](https://github.com/soerenschneider/sc/commit/afe3df4d45ea364c0ec5b0b3096cd0a7d289c119))
* update command ([538f2d4](https://github.com/soerenschneider/sc/commit/538f2d42ae52ca7e78a3b45e8bc46a7b1f27f347))

## [1.3.0](https://github.com/soerenschneider/sc/compare/v1.2.2...v1.3.0) (2025-05-18)


### Features

* add basic support for Vault ([9af30b4](https://github.com/soerenschneider/sc/commit/9af30b4815490885f3a4756d564d00ec7588779a))
* add healthcheck ([8aa9c50](https://github.com/soerenschneider/sc/commit/8aa9c5091622c0c30190c8c4a32f266e8402dc20))
* add subcommands to manage vault aws secret engine ([a9c479d](https://github.com/soerenschneider/sc/commit/a9c479de71b951181ba5e0f55e831e735674dc9f))
* add support for vault mfa totp ([1b84bd6](https://github.com/soerenschneider/sc/commit/1b84bd6ac3f3f1f3974c62f21bb2771f375e6daf))


### Bug Fixes

* **deps:** bump github.com/go-jose/go-jose/v4 from 4.0.1 to 4.0.5 ([0eae9c3](https://github.com/soerenschneider/sc/commit/0eae9c3450ed55aac1359fe3836d16f2620d23e2))
* **deps:** bump github.com/go-jose/go-jose/v4 from 4.0.1 to 4.0.5 ([1c21e0c](https://github.com/soerenschneider/sc/commit/1c21e0cac496bdb5fa1f4bbd56b2a4536ae291ff))
* **deps:** bump github.com/jedib0t/go-pretty/v6 from 6.6.5 to 6.6.7 ([39f5fb4](https://github.com/soerenschneider/sc/commit/39f5fb4431dd0a24f2a786a5015a1570226918ea))
* **deps:** bump github.com/jedib0t/go-pretty/v6 from 6.6.5 to 6.6.7 ([cd4f053](https://github.com/soerenschneider/sc/commit/cd4f0538191deee953b18f97d8a945cf9bf844c1))
* **deps:** bump github.com/soerenschneider/sc-agent from 1.8.0 to 1.10.0 ([aa2ec30](https://github.com/soerenschneider/sc/commit/aa2ec300f6ff68ded03114bf6ccad640edb76913))
* **deps:** bump github.com/soerenschneider/sc-agent from 1.8.0 to 1.10.0 ([d1e01d1](https://github.com/soerenschneider/sc/commit/d1e01d103879d0488ed87c2c102af902d7bd6d57))
* **deps:** bump github.com/spf13/afero from 1.11.0 to 1.12.0 ([33ed6dc](https://github.com/soerenschneider/sc/commit/33ed6dc50879ff5de1f67a7fe8b59e0187b432c8))
* **deps:** bump github.com/spf13/afero from 1.11.0 to 1.12.0 ([d424d62](https://github.com/soerenschneider/sc/commit/d424d62566034337bfe391f2bf0486d323c6034e))
* **deps:** bump github.com/spf13/cobra from 1.8.1 to 1.9.1 ([a4761e0](https://github.com/soerenschneider/sc/commit/a4761e08c5bc8fdc69fe3ec60951fac76f5e214a))
* **deps:** bump github.com/spf13/cobra from 1.8.1 to 1.9.1 ([58d7f53](https://github.com/soerenschneider/sc/commit/58d7f53e64ea3dba82a21102ec720e8d0fc0de8b))
* **deps:** bump golang from 1.23.4 to 1.23.6 ([09beb64](https://github.com/soerenschneider/sc/commit/09beb64f3e49bc8bf891f2f52e0655eac86bd54e))
* **deps:** bump golang from 1.23.4 to 1.23.6 ([11a314e](https://github.com/soerenschneider/sc/commit/11a314e40fb9cfda78f2602224c9a193fa0d0459))
* **deps:** bump golang from 1.23.6 to 1.24.3 ([7d56453](https://github.com/soerenschneider/sc/commit/7d56453f607a7130f4d7bfc55e1ce9b6fe7d26f7))
* **deps:** bump golang from 1.23.6 to 1.24.3 ([6e6e972](https://github.com/soerenschneider/sc/commit/6e6e972a666f2e45bc598f0621f97e5f9bd564ae))
* **deps:** bump golang.org/x/net from 0.33.0 to 0.38.0 ([38a4aad](https://github.com/soerenschneider/sc/commit/38a4aad5857ffd56431a7aa35eb5b67a733bb473))
* **deps:** bump golang.org/x/net from 0.33.0 to 0.38.0 ([0ede4ee](https://github.com/soerenschneider/sc/commit/0ede4ee64a9b384e1dffc1fcfe96341d8133facf))
* **deps:** bump golang.org/x/term from 0.28.0 to 0.29.0 ([9ab8f0f](https://github.com/soerenschneider/sc/commit/9ab8f0f95fd22d3641f9dda0cc53e8b3078a0cd7))
* **deps:** bump golang.org/x/term from 0.28.0 to 0.29.0 ([8809f77](https://github.com/soerenschneider/sc/commit/8809f7702e41ff3fc0ef1ae3464c9a5b0fa64695))
* fix api change ([1e06154](https://github.com/soerenschneider/sc/commit/1e061544b77c82f22bc9a54a7a1babbd572908d8))

## [1.2.2](https://github.com/soerenschneider/sc/compare/v1.2.1...v1.2.2) (2025-01-12)


### Bug Fixes

* **deps:** bump github.com/jedib0t/go-pretty/v6 from 6.6.1 to 6.6.5 ([1d742f1](https://github.com/soerenschneider/sc/commit/1d742f17f091718a337c5424afa99cbe58662123))
* **deps:** bump golang from 1.23.2 to 1.23.4 ([65756b0](https://github.com/soerenschneider/sc/commit/65756b04a7be6be2f98130996c81b0116dde98df))
* **deps:** bump golang.org/x/crypto from 0.28.0 to 0.31.0 ([c498c36](https://github.com/soerenschneider/sc/commit/c498c361ed6f1fff872517888c2e917b4c782ec4))
* **deps:** bump golang.org/x/net from 0.29.0 to 0.33.0 ([a470a44](https://github.com/soerenschneider/sc/commit/a470a44e51b6544d4459e1ff9aae95436d48be4c))
* **deps:** bump golang.org/x/term from 0.25.0 to 0.28.0 ([1772e78](https://github.com/soerenschneider/sc/commit/1772e78fcdf0e34ee442f9b76e8187c8231f44dc))

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
