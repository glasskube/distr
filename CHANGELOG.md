# Changelog

## [0.7.0](https://github.com/glasskube/cloud/compare/v0.6.1...v0.7.0) (2024-12-18)


### Features

* add entity sorting ([#179](https://github.com/glasskube/cloud/issues/179)) ([6737060](https://github.com/glasskube/cloud/commit/6737060290373577cbc4d2df7dba7adda031f2c7))
* add password reset ([#171](https://github.com/glasskube/cloud/issues/171)) ([1329d51](https://github.com/glasskube/cloud/commit/1329d512509e667aebd1d5de1a9a051132ea4135))
* only reopen dialog if aborted ([#174](https://github.com/glasskube/cloud/issues/174)) ([e7addc4](https://github.com/glasskube/cloud/commit/e7addc4766609a457c34dc892f54869efd51a5d0))


### Bug Fixes

* **frontend:** disable all action buttons for customer managed deployments ([#180](https://github.com/glasskube/cloud/issues/180)) ([007ced2](https://github.com/glasskube/cloud/commit/007ced21e8272f157574a0e4aab00a8adcf8243e))
* **ui:** guard routes by user role and redirect / depending on role ([#181](https://github.com/glasskube/cloud/issues/181)) ([d929744](https://github.com/glasskube/cloud/commit/d929744853a39382a2761f63cf0e5686f7f53045))


### Other

* update demo data ([#178](https://github.com/glasskube/cloud/issues/178)) ([5dc3fd0](https://github.com/glasskube/cloud/commit/5dc3fd08c175b425916cee9f8a994734e34c2c85))

## [0.6.1](https://github.com/glasskube/cloud/compare/v0.6.0...v0.6.1) (2024-12-18)


### Performance

* **backend:** optimize deployment targets query ([#175](https://github.com/glasskube/cloud/issues/175)) ([f300fcc](https://github.com/glasskube/cloud/commit/f300fcc54143bce7f78cad6e20674937e4e68d81))

## [0.6.0](https://github.com/glasskube/cloud/compare/v0.5.0...v0.6.0) (2024-12-18)


### Features

* **agent:** restart cloud agent ([#173](https://github.com/glasskube/cloud/issues/173)) ([cf5d667](https://github.com/glasskube/cloud/commit/cf5d66764722154feba5ef367e29c292594be803))
* **backend:** add sentry ([#169](https://github.com/glasskube/cloud/issues/169)) ([716987e](https://github.com/glasskube/cloud/commit/716987e80e9e8e2a1d5b0b7a545bf6148f1da614))
* **ui:** text search in tables ([#164](https://github.com/glasskube/cloud/issues/164)) ([4864ee0](https://github.com/glasskube/cloud/commit/4864ee0beef161c8a230f029e6b0e0ea3ac9beed))


### Bug Fixes

* **deps:** update dependency @sentry/angular to v8.46.0 ([#166](https://github.com/glasskube/cloud/issues/166)) ([8f112ac](https://github.com/glasskube/cloud/commit/8f112ac8cecad40a275341467b3cd9ad7eb2d6e4))
* **deps:** update dependency posthog-js to v1.202.2 ([#165](https://github.com/glasskube/cloud/issues/165)) ([57a4524](https://github.com/glasskube/cloud/commit/57a4524460a8771fba8221a2128bf6ef2f6bbf06))


### Other

* add docker-compose project name ([15d62ed](https://github.com/glasskube/cloud/commit/15d62ed7e0ebef213b0ef3f876a4cd79212f8f70))
* **deps:** update dependency tailwindcss to v3.4.17 ([#172](https://github.com/glasskube/cloud/issues/172)) ([9cb7c75](https://github.com/glasskube/cloud/commit/9cb7c759b6448f2263505abc4223b94c16bc8df8))
* log verification mails ([#170](https://github.com/glasskube/cloud/issues/170)) ([9272c2d](https://github.com/glasskube/cloud/commit/9272c2d2511d9eb985e6fe969837e5b80d44a576))

## [0.5.0](https://github.com/glasskube/cloud/compare/v0.4.0...v0.5.0) (2024-12-17)


### Features

* add conditionally disabling deploy button for vendors ([#147](https://github.com/glasskube/cloud/issues/147)) ([cee1f06](https://github.com/glasskube/cloud/commit/cee1f06d4e5973764ba3a76094bcd912d6f80030))
* add deleting applications, deployment targets, user accounts ([#139](https://github.com/glasskube/cloud/issues/139)) ([975ade7](https://github.com/glasskube/cloud/commit/975ade724e8d6450d15b44ae62b384adbffb6c64))
* add login error handling ([#160](https://github.com/glasskube/cloud/issues/160)) ([f5f0a41](https://github.com/glasskube/cloud/commit/f5f0a419d961becfaa2bcd8f79d186e255a9f251))
* add option to copy & paste the verification link ([e7fe821](https://github.com/glasskube/cloud/commit/e7fe8213716ba0872fd8955ccf57e8cb2e845209))
* add vendor as bcc and reply-to in customer invite mail ([#163](https://github.com/glasskube/cloud/issues/163)) ([e1cc4d8](https://github.com/glasskube/cloud/commit/e1cc4d8b99a1b51811e3c24a003017a828bce62b))
* **backend:** add db migrations ([#155](https://github.com/glasskube/cloud/issues/155)) ([87ffd6f](https://github.com/glasskube/cloud/commit/87ffd6f09744ea5b46e29cbc979f7a315e6b46ca))
* don't use a placeholder for password inputs ([66fe23a](https://github.com/glasskube/cloud/commit/66fe23a664b8ea51e6ba27577e6cb6d964fe231c))
* email verification ([#145](https://github.com/glasskube/cloud/issues/145)) ([e78be22](https://github.com/glasskube/cloud/commit/e78be22276abd99d0a9c23701f3baecb65ba1aac))
* make agent interval configurable on backend ([#154](https://github.com/glasskube/cloud/issues/154)) ([78ea860](https://github.com/glasskube/cloud/commit/78ea860d38ef4908ce6f1c9278be91c7a65f8925))
* **ui:** custom confirm dialog ([#151](https://github.com/glasskube/cloud/issues/151)) ([b14ca14](https://github.com/glasskube/cloud/commit/b14ca14252998916a039b99c78c508f03fa4e765))
* use jwt for agent requests ([#149](https://github.com/glasskube/cloud/issues/149)) ([b5329d6](https://github.com/glasskube/cloud/commit/b5329d6ca5628feccc862a2c74fbb2f2415ea950))


### Bug Fixes

* **deps:** update dependency @sentry/angular to v8.45.1 ([#150](https://github.com/glasskube/cloud/issues/150)) ([0e5c81b](https://github.com/glasskube/cloud/commit/0e5c81ba927e83ee0f906a858eae0c0d7824b795))
* **deps:** update dependency globe.gl to v2.34.4 ([#142](https://github.com/glasskube/cloud/issues/142)) ([fa66356](https://github.com/glasskube/cloud/commit/fa663564fb5d0d04a1abae04c52e6776a590ac93))
* **deps:** update dependency posthog-js to v1.200.1 ([#138](https://github.com/glasskube/cloud/issues/138)) ([50b0c4c](https://github.com/glasskube/cloud/commit/50b0c4cd8e8901e65e0e3b65edb3101e6382dd8b))
* **deps:** update dependency posthog-js to v1.200.2 ([#148](https://github.com/glasskube/cloud/issues/148)) ([cf17cb3](https://github.com/glasskube/cloud/commit/cf17cb35841faccc692172b0ab22f965ba5b4e16))
* **deps:** update dependency posthog-js to v1.201.0 ([#153](https://github.com/glasskube/cloud/issues/153)) ([2f6e751](https://github.com/glasskube/cloud/commit/2f6e7510efc57143186204c4594950d8bda5c488))
* **deps:** update dependency posthog-js to v1.201.1 ([#157](https://github.com/glasskube/cloud/issues/157)) ([d35be88](https://github.com/glasskube/cloud/commit/d35be88795e5f0d8ded291b7835993f595e95223))
* **deps:** update dependency posthog-js to v1.202.0 ([#159](https://github.com/glasskube/cloud/issues/159)) ([6ec9af5](https://github.com/glasskube/cloud/commit/6ec9af554cbd3feee3d0a81636be5ec13d93bada))
* **deps:** update dependency posthog-js to v1.202.1 ([#161](https://github.com/glasskube/cloud/issues/161)) ([ec23a25](https://github.com/glasskube/cloud/commit/ec23a2578b4cfdbdaa8b8f06e4c4d06558a8ffd2))
* **deps:** update font awesome to v6.7.2 ([#156](https://github.com/glasskube/cloud/issues/156)) ([adfc25f](https://github.com/glasskube/cloud/commit/adfc25fa85edf554d698902599146c0f2b8ddb08))
* **deps:** update module github.com/go-chi/chi/v5 to v5.2.0 ([#143](https://github.com/glasskube/cloud/issues/143)) ([701b7e3](https://github.com/glasskube/cloud/commit/701b7e306ca35abf3a2f7b9ea456e10f0046d37d))
* don't overwrite user name with empty string during token verification ([#167](https://github.com/glasskube/cloud/issues/167)) ([684bb83](https://github.com/glasskube/cloud/commit/684bb839c181e7567e3b0257ceef927ebc48a306))
* escape query params ([#146](https://github.com/glasskube/cloud/issues/146)) ([156cc1e](https://github.com/glasskube/cloud/commit/156cc1e002b03eda32cd9597e879ce2db91a66c7))
* revert posthog token ([#144](https://github.com/glasskube/cloud/issues/144)) ([9defc58](https://github.com/glasskube/cloud/commit/9defc58213b30c53f40ab523ec2c772ef67d3bfd))
* **ui:** fix wizard dialog on small screens ([#152](https://github.com/glasskube/cloud/issues/152)) ([d5b3e82](https://github.com/glasskube/cloud/commit/d5b3e820cdd2e0479485d95ea1d89469d3ab89e3))
* **ui:** show registration form errors after submit ([#162](https://github.com/glasskube/cloud/issues/162)) ([22eba8c](https://github.com/glasskube/cloud/commit/22eba8cf99fb9203ef01277c86092f39ec8bf302))


### Other

* **deps:** update axllent/mailpit docker tag to v1.21.7 ([#141](https://github.com/glasskube/cloud/issues/141)) ([1c1406e](https://github.com/glasskube/cloud/commit/1c1406ee21651ce35bedbe98b677ee057eed1eca))
* remove unused var ([3acfb20](https://github.com/glasskube/cloud/commit/3acfb209162e501d5855f6f6b98abbcbafc701a1))


### Docs

* add Getting started section to README ([#158](https://github.com/glasskube/cloud/issues/158)) ([987ca22](https://github.com/glasskube/cloud/commit/987ca220be3d6d0dfa8a2f928d4590160b437061))
* remove not needed database init for Getting Started ([dcbf3f0](https://github.com/glasskube/cloud/commit/dcbf3f0e82d69882df0dcd88f3650d510f8b10d6))

## [0.4.0](https://github.com/glasskube/cloud/compare/v0.3.0...v0.4.0) (2024-12-13)


### Features

* **backend:** use transactions where multiple database writes happen ([#129](https://github.com/glasskube/cloud/issues/129)) ([ef23a9f](https://github.com/glasskube/cloud/commit/ef23a9f2d560546150bd210e2c57c18c7b393fd5))
* change email footer ([cb45326](https://github.com/glasskube/cloud/commit/cb45326c5e8b5a0b9d65e7767197919de054fa63))
* **frontend:** add step 0 in onboarding wizard ([#136](https://github.com/glasskube/cloud/issues/136)) ([fda3126](https://github.com/glasskube/cloud/commit/fda312625abd87267db396354e13a42d599c94f6))


### Bug Fixes

* **deps:** update dependency @sentry/angular to v8.45.0 ([#134](https://github.com/glasskube/cloud/issues/134)) ([14ba8ee](https://github.com/glasskube/cloud/commit/14ba8eeb6d1446e6df99ee6699c0997e55b7c209))
* improve customer invite mail ([#132](https://github.com/glasskube/cloud/issues/132)) ([32f121e](https://github.com/glasskube/cloud/commit/32f121e2eee079d7f1c897259f8b5112740dc75b))


### Other

* **deps:** update jdx/mise-action action to v2.1.8 ([#133](https://github.com/glasskube/cloud/issues/133)) ([34c7b30](https://github.com/glasskube/cloud/commit/34c7b300a1b89ccbc13f08792aa691d13d834216))
* **frontend:** subheading to indicate user role ([#135](https://github.com/glasskube/cloud/issues/135)) ([bb316c0](https://github.com/glasskube/cloud/commit/bb316c06ebc2c1e9d6f61be34e4a31e9b046c1a5))

## [0.3.0](https://github.com/glasskube/cloud/compare/v0.2.0...v0.3.0) (2024-12-13)


### Features

* **frontend:** sentry and posthog user identification ([#111](https://github.com/glasskube/cloud/issues/111)) ([f592617](https://github.com/glasskube/cloud/commit/f5926177c13df51a33c91ecd8d01afa580d18ad2))
* **frontend:** version modal ([#122](https://github.com/glasskube/cloud/issues/122)) ([d9305f1](https://github.com/glasskube/cloud/commit/d9305f10e6c7f807ec58c072c60851fd50a3a9de))
* switch table header ordering and placeholders ([#121](https://github.com/glasskube/cloud/issues/121)) ([3d9f5c3](https://github.com/glasskube/cloud/commit/3d9f5c3d12c5d13bb15875d40be47fc8237ad2ea))
* **ui:** update onboarding flow to create customer user ([#123](https://github.com/glasskube/cloud/issues/123)) ([0293620](https://github.com/glasskube/cloud/commit/029362039a0dc3e6dcd336d1749719f1778aff8a))


### Bug Fixes

* **deps:** update angular monorepo to v19.0.4 ([#116](https://github.com/glasskube/cloud/issues/116)) ([f374746](https://github.com/glasskube/cloud/commit/f3747469b3dc094dedebe1fa963fe618001b095c))
* **deps:** update dependency @angular/cdk to v19.0.3 ([#106](https://github.com/glasskube/cloud/issues/106)) ([b7cf95f](https://github.com/glasskube/cloud/commit/b7cf95f3bafd141f76ee67a3792fc64898157c1d))
* **deps:** update dependency @sentry/angular to v8.44.0 ([#112](https://github.com/glasskube/cloud/issues/112)) ([3ec3176](https://github.com/glasskube/cloud/commit/3ec3176a174dc8e855817565e1a01bd8029f3a32))
* **deps:** update dependency globe.gl to v2.34.3 ([#127](https://github.com/glasskube/cloud/issues/127)) ([e7481cf](https://github.com/glasskube/cloud/commit/e7481cf6167b03e9620bbda33ecfc4b56e2a3ef7))
* **deps:** update dependency posthog-js to v1.196.1 ([#114](https://github.com/glasskube/cloud/issues/114)) ([56741a4](https://github.com/glasskube/cloud/commit/56741a4ac741daad6c0ddd2f7d60cb6b2ad63aa4))
* **deps:** update dependency posthog-js to v1.198.0 ([#118](https://github.com/glasskube/cloud/issues/118)) ([d5c3b1b](https://github.com/glasskube/cloud/commit/d5c3b1b9d4c941fa6d90e3e96735da462d267e17))
* **deps:** update dependency posthog-js to v1.199.0 ([#126](https://github.com/glasskube/cloud/issues/126)) ([edd89de](https://github.com/glasskube/cloud/commit/edd89de48f30aa869ceb1d52617df5631c3872a1))
* **deps:** update module github.com/go-chi/jwtauth/v5 to v5.3.2 ([#110](https://github.com/glasskube/cloud/issues/110)) ([55dce33](https://github.com/glasskube/cloud/commit/55dce3327664b171d163057d2ea38990b7810dc0))
* **deps:** update module golang.org/x/crypto to v0.31.0 [security] ([#109](https://github.com/glasskube/cloud/issues/109)) ([2b73573](https://github.com/glasskube/cloud/commit/2b73573129ac7cb1da8f049463955185f06fbb7b))
* **frontend:** success feedback ([#128](https://github.com/glasskube/cloud/issues/128)) ([29d9de9](https://github.com/glasskube/cloud/commit/29d9de94ace771b01ef919b64293e668bc9a1bdb))
* **frontend:** use wizard for new deployments ([#124](https://github.com/glasskube/cloud/issues/124)) ([a8b4bfa](https://github.com/glasskube/cloud/commit/a8b4bfa36e14527811a4d904224d41b3b41ab1d5))
* **ui:** add maxwith and text overflow to name columns ([#119](https://github.com/glasskube/cloud/issues/119)) ([4616a78](https://github.com/glasskube/cloud/commit/4616a781d265ea9ca6ad035e1e37b70475c78857))
* **ui:** move modal submit button and general modal improvements ([#120](https://github.com/glasskube/cloud/issues/120)) ([9b6899d](https://github.com/glasskube/cloud/commit/9b6899df3565efda55dc12e23194421f57ff2f15))


### Other

* add agent tag ([#130](https://github.com/glasskube/cloud/issues/130)) ([804eda3](https://github.com/glasskube/cloud/commit/804eda34ca3f3c1e6f4a00a4666123a66e621754))
* change 'distributor' to 'vendor' everywhere ([#115](https://github.com/glasskube/cloud/issues/115)) ([ea7ed57](https://github.com/glasskube/cloud/commit/ea7ed57db968a767b0d978f3009813c258bebc75))
* **deps:** update angular-cli monorepo to v19.0.5 ([#125](https://github.com/glasskube/cloud/issues/125)) ([e027fd6](https://github.com/glasskube/cloud/commit/e027fd6d78b5bc8773e2f1f39c9be5545ee23e86))
* **ui:** consistent naming ([#117](https://github.com/glasskube/cloud/issues/117)) ([51487f2](https://github.com/glasskube/cloud/commit/51487f2f5c4bd484228f1fdd8c39ec9877dda6de))

## [0.2.0](https://github.com/glasskube/cloud/compare/v0.1.0...v0.2.0) (2024-12-11)


### Features

* add user creation and inviting customers ([#103](https://github.com/glasskube/cloud/issues/103)) ([c2c1d8c](https://github.com/glasskube/cloud/commit/c2c1d8c2dc9ae3d2f288b1a5d576dc193b8b842b))
* customer installation wizard, dashboard charts ([#82](https://github.com/glasskube/cloud/issues/82)) ([cfa74c6](https://github.com/glasskube/cloud/commit/cfa74c67d338e84e5c00e338f00b90f7109fb8ca))


### Bug Fixes

* **frontend:** toast error message should not be 'OK' ([#104](https://github.com/glasskube/cloud/issues/104)) ([10f7a27](https://github.com/glasskube/cloud/commit/10f7a2759021ea82a33ac977a5c3e981d440e838))

## 0.1.0 (2024-12-11)


### Features

* add create application endpoint ([#15](https://github.com/glasskube/cloud/issues/15)) ([fc1f81e](https://github.com/glasskube/cloud/commit/fc1f81ee2229c2c282387a403630f2b1f804e1c4))
* add db tables and load pgx custom types ([#20](https://github.com/glasskube/cloud/issues/20)) ([5678806](https://github.com/glasskube/cloud/commit/5678806573e4186e2cac6b1c58359cfc141bc6e2))
* add foreign key indices and more dummy data ([#49](https://github.com/glasskube/cloud/issues/49)) ([ab36b2a](https://github.com/glasskube/cloud/commit/ab36b2a1a2d9d5e9649fd9cda393d9d145c3aeb6))
* add globe component ([#42](https://github.com/glasskube/cloud/issues/42)) ([1f67c2b](https://github.com/glasskube/cloud/commit/1f67c2b12ae78a55331ee1db6086bc550886b526))
* add limit to body size and file type for compose file ([#44](https://github.com/glasskube/cloud/issues/44)) ([2b1290f](https://github.com/glasskube/cloud/commit/2b1290f7d0fb1b4f9c4db52356ffea8eb50ea2bc))
* add stale status for deployment targets ([#53](https://github.com/glasskube/cloud/issues/53)) ([b48903d](https://github.com/glasskube/cloud/commit/b48903d84866d4dd0aa1b8b5ff05aab37a438ed4))
* add update application endpoint ([#17](https://github.com/glasskube/cloud/issues/17)) ([18ea0ca](https://github.com/glasskube/cloud/commit/18ea0ca547728c7efe6e0b93a8a27e415801497a))
* add user authentication ([#59](https://github.com/glasskube/cloud/issues/59)) ([d77dd22](https://github.com/glasskube/cloud/commit/d77dd221d1b75a9233650fd293c1bff484e8b451))
* add user role ([#94](https://github.com/glasskube/cloud/issues/94)) ([692395a](https://github.com/glasskube/cloud/commit/692395a4a4fe51fdf87edb1a6f1b4e433b251af6))
* agent reports status ([#93](https://github.com/glasskube/cloud/issues/93)) ([51ada87](https://github.com/glasskube/cloud/commit/51ada87221daeb9aa0fe00da7d73f9693fd7fdcb))
* application versions endpoints ([#27](https://github.com/glasskube/cloud/issues/27)) ([f2347bf](https://github.com/glasskube/cloud/commit/f2347bf852ed184057edf2c97e7543e92f657fd1))
* **backend:** add deployment target api ([#34](https://github.com/glasskube/cloud/issues/34)) ([4932d69](https://github.com/glasskube/cloud/commit/4932d69f6e78ec007a1462601e2995d5633028d6))
* **backend:** add mail sending capabilities ([#84](https://github.com/glasskube/cloud/issues/84)) ([1180e55](https://github.com/glasskube/cloud/commit/1180e55fddd704cfe887c4c825e4cadb23518d10))
* **backend:** GET /applications ([#13](https://github.com/glasskube/cloud/issues/13)) ([003f808](https://github.com/glasskube/cloud/commit/003f80876a5aa25fde83f686b6088f63ecd50288))
* cloud agent ([#83](https://github.com/glasskube/cloud/issues/83)) ([570cb1e](https://github.com/glasskube/cloud/commit/570cb1ed16c7cc04df0434541cb0b02931373eda))
* **cloud-ui:** add color scheme switcher ([#19](https://github.com/glasskube/cloud/issues/19)) ([51d2907](https://github.com/glasskube/cloud/commit/51d29077c45246d6ff47784c23152cf7cfc9071c))
* **cloud-ui:** add initial flowbite dashboard layout ([#5](https://github.com/glasskube/cloud/issues/5)) ([e6916fc](https://github.com/glasskube/cloud/commit/e6916fc7770d5ca704370bad909d37370f114fd0))
* deploy applications to deployment targets ([#54](https://github.com/glasskube/cloud/issues/54)) ([85ab2b4](https://github.com/glasskube/cloud/commit/85ab2b412652eff54f6d0d5c81198cffb2ac0ba4))
* **frontend:** form validations ([#95](https://github.com/glasskube/cloud/issues/95)) ([37564d0](https://github.com/glasskube/cloud/commit/37564d0743bc0ad32804fc6088b58fd1b5e4a4fb))
* **frontend:** global error handling ([#100](https://github.com/glasskube/cloud/issues/100)) ([e5b0bbc](https://github.com/glasskube/cloud/commit/e5b0bbcc9db69db686328e94a49e3351e41b9e8a))
* **frontend:** integrate sentry ([#101](https://github.com/glasskube/cloud/issues/101)) ([ad0cbbb](https://github.com/glasskube/cloud/commit/ad0cbbb08742e35a282609215b3a6002bb4aeb8e))
* manage versions ([#38](https://github.com/glasskube/cloud/issues/38)) ([e3f8ad2](https://github.com/glasskube/cloud/commit/e3f8ad24a415646e488d7da428f2b281cbab9109))
* migrate dropdowns to cdk overlay ([#51](https://github.com/glasskube/cloud/issues/51)) ([f222a04](https://github.com/glasskube/cloud/commit/f222a044110dab03ccf7cad2d83f04547b8f58a2))
* onboarding wizard ([#64](https://github.com/glasskube/cloud/issues/64)) ([2d3cd03](https://github.com/glasskube/cloud/commit/2d3cd034a17a2e491f33bd29e9aafa149fcf91db))
* show applications ([#16](https://github.com/glasskube/cloud/issues/16)) ([b9842c3](https://github.com/glasskube/cloud/commit/b9842c3b60b442b6d02f1e1aca92b9aa15ec984b))
* show deployment target status ([#46](https://github.com/glasskube/cloud/issues/46)) ([79cb6fa](https://github.com/glasskube/cloud/commit/79cb6fa8d03ca26c05972bd0dfd8cf8a3c2242e7))
* show deployment target status on globe ([e7002fd](https://github.com/glasskube/cloud/commit/e7002fd5a2709b3530bc4371f7df8e16ed734e5b))
* small ui improvements, add demo data sql ([#61](https://github.com/glasskube/cloud/issues/61)) ([9f257bf](https://github.com/glasskube/cloud/commit/9f257bf2725d912d68e3c6c6566cacc7c86c0129))
* support environment variables ([#58](https://github.com/glasskube/cloud/issues/58)) ([da472f2](https://github.com/glasskube/cloud/commit/da472f22b550854310b2f29634deb121775baac5))
* **ui:** add applications and deployment targets to dashboard ([#43](https://github.com/glasskube/cloud/issues/43)) ([a605e39](https://github.com/glasskube/cloud/commit/a605e39458f7ecbe2b2d1e1c1cb52a62ac6aa40b))
* **ui:** add deployment targets ui ([#36](https://github.com/glasskube/cloud/issues/36)) ([68b7b25](https://github.com/glasskube/cloud/commit/68b7b253f3f291867be7950b5746281a1f658f5b))
* **ui:** add register page ([#71](https://github.com/glasskube/cloud/issues/71)) ([4497334](https://github.com/glasskube/cloud/commit/44973348ba5ed4e0ab55dfe1cabb28fe0aac897c))
* **ui:** edit and create applications ([#25](https://github.com/glasskube/cloud/issues/25)) ([04f45d8](https://github.com/glasskube/cloud/commit/04f45d8e93e4e6c29ab2693aeeefffc6d9778f72))
* **ui:** show deployment target instructions ([#47](https://github.com/glasskube/cloud/issues/47)) ([88d6c38](https://github.com/glasskube/cloud/commit/88d6c38e85e57c55f332a1440de7d5c3c67e1a79))
* use cdk overlay for drawer overlays ([#52](https://github.com/glasskube/cloud/issues/52)) ([85e6456](https://github.com/glasskube/cloud/commit/85e6456b7d9619cbf633dda654fe7f3989b3ecc8))


### Bug Fixes

* align globe center ([#70](https://github.com/glasskube/cloud/issues/70)) ([deece53](https://github.com/glasskube/cloud/commit/deece53c252cfb45d8302dea7130405e1d7c1c79))
* always use index.html for not found files ([301d049](https://github.com/glasskube/cloud/commit/301d0497e2284a568055f270c4e129e0c8ec0442))
* **deps:** update angular monorepo to v19.0.1 ([#7](https://github.com/glasskube/cloud/issues/7)) ([fa628ee](https://github.com/glasskube/cloud/commit/fa628ee078e93edcbd12c18f6205b3e8d985935f))
* **deps:** update angular monorepo to v19.0.2 ([#72](https://github.com/glasskube/cloud/issues/72)) ([14bb184](https://github.com/glasskube/cloud/commit/14bb1843d5c36dd1b4e99e3e4d170f5adcc2770f))
* **deps:** update angular monorepo to v19.0.3 ([#75](https://github.com/glasskube/cloud/issues/75)) ([7f7679f](https://github.com/glasskube/cloud/commit/7f7679fdbf6ea82acdd665068c4c5b834945fcf4))
* **deps:** update dependency @angular/cdk to v19.0.2 ([#76](https://github.com/glasskube/cloud/issues/76)) ([408fbf2](https://github.com/glasskube/cloud/commit/408fbf2cd9b7c59e685c221d88b38dd649730354))
* **deps:** update dependency globe.gl to v2.34.2 ([#57](https://github.com/glasskube/cloud/issues/57)) ([f3a0265](https://github.com/glasskube/cloud/commit/f3a02653c2297dc90d1f4b465e0aaba57caddb42))
* **deps:** update dependency posthog-js to v1.188.0 ([#24](https://github.com/glasskube/cloud/issues/24)) ([afa94e7](https://github.com/glasskube/cloud/commit/afa94e739e8ba314791f1409bb8e3706507a44c1))
* **deps:** update dependency posthog-js to v1.188.1 ([#29](https://github.com/glasskube/cloud/issues/29)) ([67e7448](https://github.com/glasskube/cloud/commit/67e74485d95cce3fbd3443a054f6a910023bfa6c))
* **deps:** update dependency posthog-js to v1.189.0 ([#37](https://github.com/glasskube/cloud/issues/37)) ([b6c00c5](https://github.com/glasskube/cloud/commit/b6c00c5556d39e0dc07f2ddab077d7fe62688aba))
* **deps:** update dependency posthog-js to v1.190.1 ([#39](https://github.com/glasskube/cloud/issues/39)) ([87358d6](https://github.com/glasskube/cloud/commit/87358d69e095ab0fdc63219490706357584c04bf))
* **deps:** update dependency posthog-js to v1.190.2 ([#40](https://github.com/glasskube/cloud/issues/40)) ([fc1c433](https://github.com/glasskube/cloud/commit/fc1c433acf19429a4308eb779085ce0401c53715))
* **deps:** update dependency posthog-js to v1.191.0 ([#45](https://github.com/glasskube/cloud/issues/45)) ([255ca0a](https://github.com/glasskube/cloud/commit/255ca0a83d6cbed5f5260f7cbb549821aae29ec5))
* **deps:** update dependency posthog-js to v1.192.1 ([#50](https://github.com/glasskube/cloud/issues/50)) ([1b00f6c](https://github.com/glasskube/cloud/commit/1b00f6c9cb4b445d3f87b477b44d1b4ea355a809))
* **deps:** update dependency posthog-js to v1.193.1 ([#55](https://github.com/glasskube/cloud/issues/55)) ([dfe3cb8](https://github.com/glasskube/cloud/commit/dfe3cb8538030a2baba0b32f4448146b27a64390))
* **deps:** update dependency posthog-js to v1.194.1 ([#56](https://github.com/glasskube/cloud/issues/56)) ([6eecdec](https://github.com/glasskube/cloud/commit/6eecdecbdd641d21833b5c7251c4321af383b358))
* **deps:** update dependency posthog-js to v1.194.2 ([#60](https://github.com/glasskube/cloud/issues/60)) ([c145062](https://github.com/glasskube/cloud/commit/c1450628dc512df68125cca73ba1aaaad5b798b1))
* **deps:** update dependency posthog-js to v1.194.3 ([#63](https://github.com/glasskube/cloud/issues/63)) ([89f1928](https://github.com/glasskube/cloud/commit/89f1928c50e0a79880d0e7b8022ab89508c6a872))
* **deps:** update dependency posthog-js to v1.194.4 ([#85](https://github.com/glasskube/cloud/issues/85)) ([f47bbb5](https://github.com/glasskube/cloud/commit/f47bbb5ef8c02a71a7bb82a869831639faf6412a))
* **deps:** update dependency posthog-js to v1.194.5 ([#87](https://github.com/glasskube/cloud/issues/87)) ([f8d4dc0](https://github.com/glasskube/cloud/commit/f8d4dc0d4b8007adc82277759f39fa8d19a90236))
* **deps:** update dependency posthog-js to v1.194.6 ([#92](https://github.com/glasskube/cloud/issues/92)) ([524dec4](https://github.com/glasskube/cloud/commit/524dec4f9b96dd821d56b7f566a518d27a1258e6))
* **deps:** update dependency posthog-js to v1.195.0 ([#102](https://github.com/glasskube/cloud/issues/102)) ([4ce0930](https://github.com/glasskube/cloud/commit/4ce0930dd31f410467f20fc703f689fd0b41608c))
* **deps:** update module github.com/lestrrat-go/jwx/v2 to v2.0.21 [security] ([#68](https://github.com/glasskube/cloud/issues/68)) ([f4042fd](https://github.com/glasskube/cloud/commit/f4042fd781b15a2b2754193dce39984511a5490f))
* **deps:** update module github.com/lestrrat-go/jwx/v2 to v2.1.3 ([#73](https://github.com/glasskube/cloud/issues/73)) ([9c65205](https://github.com/glasskube/cloud/commit/9c65205798d343ee3d816754093032002431fd82))
* **deps:** update module github.com/onsi/gomega to v1.36.1 ([#97](https://github.com/glasskube/cloud/issues/97)) ([133a623](https://github.com/glasskube/cloud/commit/133a623b1b483219033037643375e1a779d7dfe7))
* **deps:** update module golang.org/x/crypto to v0.30.0 ([#77](https://github.com/glasskube/cloud/issues/77)) ([3d8d925](https://github.com/glasskube/cloud/commit/3d8d9259e2c744f56b99adfb47201f1140641561))
* **frontend:** logout if token expired or 401 response ([#99](https://github.com/glasskube/cloud/issues/99)) ([fac37da](https://github.com/glasskube/cloud/commit/fac37da215b46ffc2afa0cdac8fd18c127c78e21))
* toggle sidebar on tiny screens ([#96](https://github.com/glasskube/cloud/issues/96)) ([0e31121](https://github.com/glasskube/cloud/commit/0e3112185754d25bc953d5c54aef808cc1a62071))
* **ui:** application cache ([#41](https://github.com/glasskube/cloud/issues/41)) ([a50fb26](https://github.com/glasskube/cloud/commit/a50fb26ad17d3a9d1a165bb290a310dd91eedcd9))


### Other

* add basic go app with file server ([3a45ef9](https://github.com/glasskube/cloud/commit/3a45ef9a6b10ba528a6a4200528ac3b42353c440))
* add CHANGELOG.md to prettierignore file ([8da206b](https://github.com/glasskube/cloud/commit/8da206b851bad3514bd7b3a9e460219f6e0aec73))
* add dockerfile ([efc29b5](https://github.com/glasskube/cloud/commit/efc29b58352d988eed295d85140f76ee8c9a85f4))
* add golangci-lint ([622bf04](https://github.com/glasskube/cloud/commit/622bf04c4cf696328dbe77bad92b8b03104ebcf8))
* add missing fonts ([#21](https://github.com/glasskube/cloud/issues/21)) ([e114de9](https://github.com/glasskube/cloud/commit/e114de917d5dbe77e88a24f51fff4a8b1033a9b9))
* add posthog-js ([#8](https://github.com/glasskube/cloud/issues/8)) ([32b1ad0](https://github.com/glasskube/cloud/commit/32b1ad0ecb06eb57570b706144cabd4ab015e8cf))
* add prettier ([#11](https://github.com/glasskube/cloud/issues/11)) ([8c91f1d](https://github.com/glasskube/cloud/commit/8c91f1d35ec0952a0a7fe05f8070b57ee96201d7))
* add proxy config for dev ([449eb7f](https://github.com/glasskube/cloud/commit/449eb7f833ec20b920eafb366de2897cd560a047))
* add renovate automerge config ([4b3887a](https://github.com/glasskube/cloud/commit/4b3887a03fd0bbd0eef776ec529a3cbfb8ceecf7))
* add routing with chi ([090536f](https://github.com/glasskube/cloud/commit/090536f64b57752208e08320f8d1206a13603f36))
* add tailwind ([b5ee219](https://github.com/glasskube/cloud/commit/b5ee2194756b0211699ddb778c41f5bed1183ed8))
* change docker base to chainguard ([7457a87](https://github.com/glasskube/cloud/commit/7457a87102e404f139b56196712061817113e0d7))
* change fs embed path ([c0122eb](https://github.com/glasskube/cloud/commit/c0122eb66638e609b485cc670e77fa7800bb604c))
* configure Renovate ([43a2617](https://github.com/glasskube/cloud/commit/43a26179dcecb419f0fef9bb7571a7b5f1fd62c3))
* create angular app ([aedf40c](https://github.com/glasskube/cloud/commit/aedf40c85797f79cb136a21b65fea4e841d1b045))
* delete duplicate html ([69ea918](https://github.com/glasskube/cloud/commit/69ea918ead9a52e9423f3065d0729cba76caa5a9))
* **deps:** update Angular to v19.0.0 ([45db8c9](https://github.com/glasskube/cloud/commit/45db8c9bbcb5f86da5372c71b9e82aff8ba56eb7))
* **deps:** update angular-cli monorepo to v19.0.1 ([#18](https://github.com/glasskube/cloud/issues/18)) ([50a8c4e](https://github.com/glasskube/cloud/commit/50a8c4ef67bacb00514a00ddb8f4ce2cc2b6b65c))
* **deps:** update angular-cli monorepo to v19.0.2 ([#26](https://github.com/glasskube/cloud/issues/26)) ([136fb8f](https://github.com/glasskube/cloud/commit/136fb8fe694ec92e581f8fd90f82175771795822))
* **deps:** update angular-cli monorepo to v19.0.3 ([#74](https://github.com/glasskube/cloud/issues/74)) ([44a0551](https://github.com/glasskube/cloud/commit/44a0551a34e244d0bcf6df8c4cf60a734255d138))
* **deps:** update angular-cli monorepo to v19.0.4 ([#86](https://github.com/glasskube/cloud/issues/86)) ([361e448](https://github.com/glasskube/cloud/commit/361e448cbd6205c0c3c6d54a38d16ef0e656baa4))
* **deps:** update axllent/mailpit docker tag to v1.21.5 ([#88](https://github.com/glasskube/cloud/issues/88)) ([c583e74](https://github.com/glasskube/cloud/commit/c583e74da60d94354db6f26865d32d8f91b9e78e))
* **deps:** update axllent/mailpit docker tag to v1.21.6 ([#91](https://github.com/glasskube/cloud/issues/91)) ([b1d694c](https://github.com/glasskube/cloud/commit/b1d694c61a9fec0254911a7d19be0981f93651d6))
* **deps:** update cgr.dev/chainguard/static:latest docker digest to 5ff428f ([#14](https://github.com/glasskube/cloud/issues/14)) ([aab5ae8](https://github.com/glasskube/cloud/commit/aab5ae8e22c9f850c38e5498d43423fc1e2ac9b2))
* **deps:** update dependency @types/jasmine to v5.1.5 ([#48](https://github.com/glasskube/cloud/issues/48)) ([056df29](https://github.com/glasskube/cloud/commit/056df296ee25664672904434a57d0d76e863df4f))
* **deps:** update dependency jasmine-core to ~5.4.0 ([#2](https://github.com/glasskube/cloud/issues/2)) ([f4123f4](https://github.com/glasskube/cloud/commit/f4123f4011b53e02c750e761165d60487d8c930d))
* **deps:** update dependency jasmine-core to ~5.5.0 ([#62](https://github.com/glasskube/cloud/issues/62)) ([e5c43d9](https://github.com/glasskube/cloud/commit/e5c43d9e60b24fb1afb8b9de92fe271256001aac))
* **deps:** update dependency prettier to v3.4.0 ([#28](https://github.com/glasskube/cloud/issues/28)) ([198dc31](https://github.com/glasskube/cloud/commit/198dc31114c1155294172b4b058847e2043bb9b8))
* **deps:** update dependency prettier to v3.4.1 ([#35](https://github.com/glasskube/cloud/issues/35)) ([7cf4e6c](https://github.com/glasskube/cloud/commit/7cf4e6c595c582442c20a8a3c22e3d692f748a56))
* **deps:** update dependency prettier to v3.4.2 ([#67](https://github.com/glasskube/cloud/issues/67)) ([11db641](https://github.com/glasskube/cloud/commit/11db6415e65a7330e27a48bdc05687273c6c46bb))
* **deps:** update dependency tailwindcss to v3.4.16 ([#65](https://github.com/glasskube/cloud/issues/65)) ([5732d9f](https://github.com/glasskube/cloud/commit/5732d9f29566997dcf25de7dc189f9b4b79ea183))
* **deps:** update dependency typescript to ~5.6.0 ([#3](https://github.com/glasskube/cloud/issues/3)) ([de8296a](https://github.com/glasskube/cloud/commit/de8296a3fd0e9939ba66e1462d60f21592b48e24))
* **deps:** update node.js to v22.11.0 ([#6](https://github.com/glasskube/cloud/issues/6)) ([af79add](https://github.com/glasskube/cloud/commit/af79add744c4b344706d89eb68c1a779fc603322))
* **deps:** update node.js to v22.12.0 ([#66](https://github.com/glasskube/cloud/issues/66)) ([575d176](https://github.com/glasskube/cloud/commit/575d1763266b76faaef4b0763d804b1e57827331))
* **deps:** update postgres docker tag to v17.2 ([#23](https://github.com/glasskube/cloud/issues/23)) ([eef377c](https://github.com/glasskube/cloud/commit/eef377c541649b9e6e2df6f3379621e8ffcdcfa7))
* init angular ([a3450af](https://github.com/glasskube/cloud/commit/a3450af130564068110ee346abb8fcec2eed85f2))
* rename frontend to cloud-ui ([a685e98](https://github.com/glasskube/cloud/commit/a685e985b209d475b835764aee4bdcaa7409f2cf))
* run go mod tidy ([9f53a60](https://github.com/glasskube/cloud/commit/9f53a60fb20d911dafd30b565eb16c9c23c0fcc5))
* set angular analytics to false ([0c9b912](https://github.com/glasskube/cloud/commit/0c9b912e8aae9dc3cc1ec48c8a52f8cd908bc3ca))
* set next release to 0.1.0 ([f0a6d3e](https://github.com/glasskube/cloud/commit/f0a6d3e87c23098a3f7d07c214a65d10f5a7a6b5))
* update dummy data ([3acd07e](https://github.com/glasskube/cloud/commit/3acd07e4b89f8cf194d117df5d4e8fa254077278))


### Docs

* remove outdated information from README ([fb05128](https://github.com/glasskube/cloud/commit/fb051280c48ae6890ec7a66e2713010790280c93))


### Refactoring

* **cloud-ui:** use *ngIf to display alerts, don't use flowbite-angular, remove duplicate code ([#12](https://github.com/glasskube/cloud/issues/12)) ([fca2bc2](https://github.com/glasskube/cloud/commit/fca2bc2ddb40fd4f28ec59b1c5e73005eb0c5abf))
* reorganize backend modules to better separate service init, routing and serving ([#90](https://github.com/glasskube/cloud/issues/90)) ([f9dd232](https://github.com/glasskube/cloud/commit/f9dd232afef547a4242bd4349383a1dfae65a683))
