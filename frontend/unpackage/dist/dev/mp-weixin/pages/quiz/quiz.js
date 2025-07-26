(global["webpackJsonp"] = global["webpackJsonp"] || []).push([["pages/quiz/quiz"],{

/***/ 100:
/*!********************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
  !*** ./node_modules/babel-loader/lib!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-preprocess-loader??ref--13-1!./node_modules/@dcloudio/webpack-uni-mp-loader/lib/script.js!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib??vue-loader-options!./node_modules/@dcloudio/webpack-uni-mp-loader/lib/style.js!/Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue?vue&type=script&lang=js& ***!
  \********************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/* WEBPACK VAR INJECTION */(function(uni) {

var _interopRequireDefault = __webpack_require__(/*! @babel/runtime/helpers/interopRequireDefault */ 4);
Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;
var _regenerator = _interopRequireDefault(__webpack_require__(/*! @babel/runtime/regenerator */ 88));
var _defineProperty2 = _interopRequireDefault(__webpack_require__(/*! @babel/runtime/helpers/defineProperty */ 11));
var _slicedToArray2 = _interopRequireDefault(__webpack_require__(/*! @babel/runtime/helpers/slicedToArray */ 5));
var _asyncToGenerator2 = _interopRequireDefault(__webpack_require__(/*! @babel/runtime/helpers/asyncToGenerator */ 90));
var _papers = __webpack_require__(/*! @/api/papers */ 104);
var _levels = __webpack_require__(/*! @/api/levels */ 105);
var _questions = __webpack_require__(/*! @/api/questions */ 255);
var _canvasConfetti = _interopRequireDefault(__webpack_require__(/*! canvas-confetti */ 256));
function ownKeys(object, enumerableOnly) { var keys = Object.keys(object); if (Object.getOwnPropertySymbols) { var symbols = Object.getOwnPropertySymbols(object); enumerableOnly && (symbols = symbols.filter(function (sym) { return Object.getOwnPropertyDescriptor(object, sym).enumerable; })), keys.push.apply(keys, symbols); } return keys; }
function _objectSpread(target) { for (var i = 1; i < arguments.length; i++) { var source = null != arguments[i] ? arguments[i] : {}; i % 2 ? ownKeys(Object(source), !0).forEach(function (key) { (0, _defineProperty2.default)(target, key, source[key]); }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(target, Object.getOwnPropertyDescriptors(source)) : ownKeys(Object(source)).forEach(function (key) { Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key)); }); } return target; }
// 添加这行
var _default = {
  data: function data() {
    return {
      paperId: '',
      questionPairs: [],
      // [[id1, id2], [id3, id4], ...] 存储成对的问题ID
      currentPairIndex: 0,
      leadInQuestion: null,
      // 当前引入题
      conceptQuestion: null,
      // 当前概念题
      showConceptQuestion: false,
      leadInSelectedOption: null,
      conceptSelectedOption: null,
      leadInShowResult: false,
      conceptShowResult: false,
      loading: false,
      // 添加计时相关数据
      startTime: null,
      leadInStartTime: null,
      conceptStartTime: null
    };
  },
  computed: {
    currentConceptName: function currentConceptName() {
      if (!this.leadInQuestion) return '';
      try {
        return JSON.parse(this.leadInQuestion.content_json).concept_name;
      } catch (e) {
        return '';
      }
    },
    leadInQuestionContent: function leadInQuestionContent() {
      if (!this.leadInQuestion) return {};
      try {
        return JSON.parse(this.leadInQuestion.content_json);
      } catch (e) {
        return {};
      }
    },
    leadInAnswerContent: function leadInAnswerContent() {
      if (!this.leadInQuestion) return {};
      try {
        return JSON.parse(this.leadInQuestion.answer_json);
      } catch (e) {
        return {};
      }
    },
    conceptQuestionContent: function conceptQuestionContent() {
      if (!this.conceptQuestion) return {};
      try {
        return JSON.parse(this.conceptQuestion.content_json);
      } catch (e) {
        return {};
      }
    },
    conceptAnswerContent: function conceptAnswerContent() {
      if (!this.conceptQuestion) return {};
      try {
        return JSON.parse(this.conceptQuestion.answer_json);
      } catch (e) {
        return {};
      }
    },
    isLeadInAnswerCorrect: function isLeadInAnswerCorrect() {
      return this.getOptionLetter(this.leadInSelectedOption) === this.leadInAnswerContent.correct_option;
    },
    isConceptAnswerCorrect: function isConceptAnswerCorrect() {
      return this.getOptionLetter(this.conceptSelectedOption) === this.conceptAnswerContent.correct_option;
    },
    nextButtonText: function nextButtonText() {
      if (!this.showConceptQuestion) {
        return this.isLeadInAnswerCorrect ? '查看概念题' : '重新作答';
      }
      return this.currentPairIndex < this.questionPairs.length - 1 ? '下一组题目' : '完成';
    }
  },
  onLoad: function onLoad(options) {
    if (options.id) {
      this.paperId = options.id;
      this.startTime = Date.now();
      this.loadLevel();
    }
  },
  methods: {
    goBack: function goBack() {
      uni.navigateBack();
    },
    loadLevel: function loadLevel() {
      var _this = this;
      return (0, _asyncToGenerator2.default)( /*#__PURE__*/_regenerator.default.mark(function _callee() {
        var levelResponse;
        return _regenerator.default.wrap(function _callee$(_context) {
          while (1) {
            switch (_context.prev = _context.next) {
              case 0:
                _this.loading = true;
                _context.prev = 1;
                _context.next = 4;
                return (0, _papers.getPaperLevel)(_this.paperId);
              case 4:
                levelResponse = _context.sent;
                if (!levelResponse.success) {
                  _context.next = 19;
                  break;
                }
                _this.level = levelResponse.data;
                _this.levelId = _this.level.id; // 保存 levelId

                // 开始关卡
                _context.prev = 8;
                _context.next = 11;
                return (0, _levels.startLevel)(_this.levelId);
              case 11:
                console.log('关卡开始成功');
                _context.next = 17;
                break;
              case 14:
                _context.prev = 14;
                _context.t0 = _context["catch"](8);
                console.error('开始关卡失败:', _context.t0);
                // 不阻止用户继续，只记录错误
              case 17:
                _context.next = 19;
                return _this.loadQuestionIds(_this.level.id);
              case 19:
                _context.next = 25;
                break;
              case 21:
                _context.prev = 21;
                _context.t1 = _context["catch"](1);
                console.error('获取关卡失败:', _context.t1);
                _this.handleApiError(_context.t1);
              case 25:
                _context.prev = 25;
                _this.loading = false;
                return _context.finish(25);
              case 28:
              case "end":
                return _context.stop();
            }
          }
        }, _callee, null, [[1, 21, 25, 28], [8, 14]]);
      }))();
    },
    loadQuestionIds: function loadQuestionIds(levelId) {
      var _this2 = this;
      return (0, _asyncToGenerator2.default)( /*#__PURE__*/_regenerator.default.mark(function _callee2() {
        var questionsResponse, allQuestions, i;
        return _regenerator.default.wrap(function _callee2$(_context2) {
          while (1) {
            switch (_context2.prev = _context2.next) {
              case 0:
                _context2.prev = 0;
                _context2.next = 3;
                return (0, _levels.getLevelQuestions)(levelId);
              case 3:
                questionsResponse = _context2.sent;
                if (!questionsResponse.success) {
                  _context2.next = 11;
                  break;
                }
                // 将问题ID两两配对
                allQuestions = questionsResponse.data;
                _this2.questionPairs = [];
                for (i = 0; i < allQuestions.length; i += 2) {
                  _this2.questionPairs.push([allQuestions[i].id, allQuestions[i + 1].id]);
                }

                // 加载第一个引入题
                if (!(_this2.questionPairs.length > 0)) {
                  _context2.next = 11;
                  break;
                }
                _context2.next = 11;
                return _this2.loadCurrentPair();
              case 11:
                _context2.next = 17;
                break;
              case 13:
                _context2.prev = 13;
                _context2.t0 = _context2["catch"](0);
                console.error('获取问题列表失败:', _context2.t0);
                _this2.handleApiError(_context2.t0);
              case 17:
              case "end":
                return _context2.stop();
            }
          }
        }, _callee2, null, [[0, 13]]);
      }))();
    },
    loadCurrentPair: function loadCurrentPair() {
      var _this3 = this;
      return (0, _asyncToGenerator2.default)( /*#__PURE__*/_regenerator.default.mark(function _callee3() {
        var _yield$Promise$all, _yield$Promise$all2, leadInResponse, conceptResponse;
        return _regenerator.default.wrap(function _callee3$(_context3) {
          while (1) {
            switch (_context3.prev = _context3.next) {
              case 0:
                _this3.loading = true;
                _context3.prev = 1;
                _context3.next = 4;
                return Promise.all([(0, _questions.getQuestion)(_this3.questionPairs[_this3.currentPairIndex][0]), (0, _questions.getQuestion)(_this3.questionPairs[_this3.currentPairIndex][1])]);
              case 4:
                _yield$Promise$all = _context3.sent;
                _yield$Promise$all2 = (0, _slicedToArray2.default)(_yield$Promise$all, 2);
                leadInResponse = _yield$Promise$all2[0];
                conceptResponse = _yield$Promise$all2[1];
                if (leadInResponse.success && conceptResponse.success) {
                  _this3.leadInQuestion = leadInResponse.data;
                  _this3.conceptQuestion = conceptResponse.data;
                  _this3.resetQuestionState();
                }
                _context3.next = 15;
                break;
              case 11:
                _context3.prev = 11;
                _context3.t0 = _context3["catch"](1);
                console.error('加载题目失败:', _context3.t0);
                _this3.handleApiError(_context3.t0);
              case 15:
                _context3.prev = 15;
                _this3.loading = false;
                return _context3.finish(15);
              case 18:
              case "end":
                return _context3.stop();
            }
          }
        }, _callee3, null, [[1, 11, 15, 18]]);
      }))();
    },
    resetQuestionState: function resetQuestionState() {
      this.showConceptQuestion = false;
      this.leadInSelectedOption = null;
      this.conceptSelectedOption = null;
      this.leadInShowResult = false;
      this.conceptShowResult = false;
      // 重置计时器
      this.leadInStartTime = Date.now();
    },
    // 添加这两个新方法
    fireBasicConfetti: function fireBasicConfetti() {
      (0, _canvasConfetti.default)({
        particleCount: 100,
        spread: 70,
        origin: {
          y: 0.6
        }
      });
    },
    fireSchoolPride: function fireSchoolPride() {
      function fire(particleRatio, opts) {
        (0, _canvasConfetti.default)(_objectSpread(_objectSpread({}, opts), {}, {
          origin: {
            y: 0.7
          },
          particleCount: Math.floor(200 * particleRatio)
        }));
      }
      fire(0.25, {
        spread: 26,
        startVelocity: 55,
        origin: {
          x: 0.2
        }
      });
      fire(0.25, {
        spread: 26,
        startVelocity: 55,
        origin: {
          x: 0.8
        }
      });
      setTimeout(function () {
        fire(0.2, {
          spread: 60,
          origin: {
            x: 0.2
          }
        });
        fire(0.2, {
          spread: 60,
          origin: {
            x: 0.8
          }
        });
      }, 150);
    },
    // 修改 selectLeadInOption 方法
    selectLeadInOption: function selectLeadInOption(index) {
      var _this4 = this;
      return (0, _asyncToGenerator2.default)( /*#__PURE__*/_regenerator.default.mark(function _callee4() {
        var isCorrect, durationMs, answerJson;
        return _regenerator.default.wrap(function _callee4$(_context4) {
          while (1) {
            switch (_context4.prev = _context4.next) {
              case 0:
                if (!_this4.leadInShowResult) {
                  _context4.next = 2;
                  break;
                }
                return _context4.abrupt("return");
              case 2:
                _this4.leadInSelectedOption = index;
                _this4.leadInShowResult = true;
                isCorrect = _this4.isLeadInOptionCorrect(index);
                durationMs = Date.now() - (_this4.leadInStartTime || _this4.startTime); // 提交答案 - 修改 answer_json 格式
                _context4.prev = 6;
                answerJson = {
                  type: "multiple_choice",
                  selected_option: _this4.getOptionLetter(index),
                  correct_option: _this4.leadInAnswerContent.correct_option,
                  is_correct: isCorrect
                };
                _context4.next = 10;
                return (0, _levels.submitAnswer)(_this4.levelId, _this4.leadInQuestion.id, answerJson, durationMs);
              case 10:
                _context4.next = 15;
                break;
              case 12:
                _context4.prev = 12;
                _context4.t0 = _context4["catch"](6);
                console.error('提交答案失败:', _context4.t0);
                // 不阻止用户继续，只记录错误
              case 15:
                if (isCorrect) {
                  _this4.showConceptQuestion = true;
                  _this4.conceptStartTime = Date.now();
                } else {
                  uni.showToast({
                    title: '答案错误',
                    icon: 'none'
                  });
                }
              case 16:
              case "end":
                return _context4.stop();
            }
          }
        }, _callee4, null, [[6, 12]]);
      }))();
    },
    // 修改 selectConceptOption 方法
    selectConceptOption: function selectConceptOption(index) {
      var _this5 = this;
      return (0, _asyncToGenerator2.default)( /*#__PURE__*/_regenerator.default.mark(function _callee5() {
        var isCorrect, durationMs, answerJson;
        return _regenerator.default.wrap(function _callee5$(_context5) {
          while (1) {
            switch (_context5.prev = _context5.next) {
              case 0:
                if (!_this5.conceptShowResult) {
                  _context5.next = 2;
                  break;
                }
                return _context5.abrupt("return");
              case 2:
                _this5.conceptSelectedOption = index;
                _this5.conceptShowResult = true;
                isCorrect = _this5.isConceptOptionCorrect(index);
                durationMs = Date.now() - (_this5.conceptStartTime || _this5.startTime); // 提交答案 - 修改 answer_json 格式
                _context5.prev = 6;
                answerJson = {
                  type: "multiple_choice",
                  selected_option: _this5.getOptionLetter(index),
                  correct_option: _this5.conceptAnswerContent.correct_option,
                  is_correct: isCorrect
                };
                _context5.next = 10;
                return (0, _levels.submitAnswer)(_this5.levelId, _this5.conceptQuestion.id, answerJson, durationMs);
              case 10:
                _context5.next = 15;
                break;
              case 12:
                _context5.prev = 12;
                _context5.t0 = _context5["catch"](6);
                console.error('提交答案失败:', _context5.t0);
                // 不阻止用户继续，只记录错误
              case 15:
                if (isCorrect) {
                  _this5.fireBasicConfetti();
                } else {
                  uni.showToast({
                    title: '答案错误',
                    icon: 'none'
                  });
                }
              case 16:
              case "end":
                return _context5.stop();
            }
          }
        }, _callee5, null, [[6, 12]]);
      }))();
    },
    // 修改 handleNext 方法
    handleNext: function handleNext() {
      var _this6 = this;
      return (0, _asyncToGenerator2.default)( /*#__PURE__*/_regenerator.default.mark(function _callee6() {
        return _regenerator.default.wrap(function _callee6$(_context6) {
          while (1) {
            switch (_context6.prev = _context6.next) {
              case 0:
                if (_this6.showConceptQuestion) {
                  _context6.next = 4;
                  break;
                }
                // 引入题部分
                if (_this6.isLeadInAnswerCorrect) {
                  _this6.showConceptQuestion = true;
                  _this6.conceptStartTime = Date.now();
                } else {
                  _this6.leadInSelectedOption = null;
                  _this6.leadInShowResult = false;
                }
                _context6.next = 21;
                break;
              case 4:
                if (!_this6.conceptShowResult) {
                  _context6.next = 21;
                  break;
                }
                if (!(_this6.currentPairIndex < _this6.questionPairs.length - 1)) {
                  _context6.next = 10;
                  break;
                }
                _this6.currentPairIndex++;
                _this6.loadCurrentPair();
                _context6.next = 21;
                break;
              case 10:
                _context6.prev = 10;
                _context6.next = 13;
                return (0, _levels.completeLevel)(_this6.levelId);
              case 13:
                console.log('关卡完成提交成功');
                _context6.next = 19;
                break;
              case 16:
                _context6.prev = 16;
                _context6.t0 = _context6["catch"](10);
                console.error('提交关卡完成失败:', _context6.t0);
                // 不阻止用户继续，只记录错误
              case 19:
                _this6.fireSchoolPride();
                uni.showToast({
                  title: '恭喜完成所有题目！',
                  icon: 'success',
                  duration: 2000,
                  complete: function complete() {
                    setTimeout(function () {
                      return uni.navigateBack();
                    }, 2000);
                  }
                });
              case 21:
              case "end":
                return _context6.stop();
            }
          }
        }, _callee6, null, [[10, 16]]);
      }))();
    },
    getOptionLetter: function getOptionLetter(index) {
      return String.fromCharCode(65 + index);
    },
    isLeadInOptionCorrect: function isLeadInOptionCorrect(index) {
      return this.getOptionLetter(index) === this.leadInAnswerContent.correct_option;
    },
    isConceptOptionCorrect: function isConceptOptionCorrect(index) {
      return this.getOptionLetter(index) === this.conceptAnswerContent.correct_option;
    },
    handleApiError: function handleApiError(error) {
      console.error('API错误:', error);
      if (error.message === '请重新登录') {
        uni.removeStorageSync('access_token');
        uni.removeStorageSync('refresh_token');
        uni.removeStorageSync('token_expires_in');
        uni.removeStorageSync('userInfo');
        uni.showToast({
          title: '请重新登录',
          icon: 'none',
          duration: 2000,
          complete: function complete() {
            setTimeout(function () {
              uni.reLaunch({
                url: '/pages/index/index'
              });
            }, 1000);
          }
        });
      } else {
        uni.showToast({
          title: error.message || '发生错误',
          icon: 'none'
        });
      }
    }
  }
};
exports.default = _default;
/* WEBPACK VAR INJECTION */}.call(this, __webpack_require__(/*! ./node_modules/@dcloudio/uni-mp-weixin/dist/index.js */ 2)["default"]))

/***/ }),

/***/ 101:
/*!*********************************************************************************************************************!*\
  !*** /Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue?vue&type=style&index=0&lang=css& ***!
  \*********************************************************************************************************************/
/*! no static exports found */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_stylePostLoader_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_6_oneOf_1_2_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_3_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/mini-css-extract-plugin/dist/loader.js??ref--6-oneOf-1-0!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib/loaders/stylePostLoader.js!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-preprocess-loader??ref--6-oneOf-1-2!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/postcss-loader/src??ref--6-oneOf-1-3!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib??vue-loader-options!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/webpack-uni-mp-loader/lib/style.js!./quiz.vue?vue&type=style&index=0&lang=css& */ 102);
/* harmony import */ var _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_stylePostLoader_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_6_oneOf_1_2_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_3_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0___default = /*#__PURE__*/__webpack_require__.n(_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_stylePostLoader_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_6_oneOf_1_2_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_3_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__);
/* harmony reexport (unknown) */ for(var __WEBPACK_IMPORT_KEY__ in _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_stylePostLoader_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_6_oneOf_1_2_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_3_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__) if(["default"].indexOf(__WEBPACK_IMPORT_KEY__) < 0) (function(key) { __webpack_require__.d(__webpack_exports__, key, function() { return _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_stylePostLoader_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_6_oneOf_1_2_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_3_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0__[key]; }) }(__WEBPACK_IMPORT_KEY__));
 /* harmony default export */ __webpack_exports__["default"] = (_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_mini_css_extract_plugin_dist_loader_js_ref_6_oneOf_1_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_css_loader_dist_cjs_js_ref_6_oneOf_1_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_stylePostLoader_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_6_oneOf_1_2_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_postcss_loader_src_index_js_ref_6_oneOf_1_3_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_0___default.a); 

/***/ }),

/***/ 102:
/*!*************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
  !*** ./node_modules/mini-css-extract-plugin/dist/loader.js??ref--6-oneOf-1-0!./node_modules/css-loader/dist/cjs.js??ref--6-oneOf-1-1!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib/loaders/stylePostLoader.js!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-preprocess-loader??ref--6-oneOf-1-2!./node_modules/postcss-loader/src??ref--6-oneOf-1-3!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib??vue-loader-options!./node_modules/@dcloudio/webpack-uni-mp-loader/lib/style.js!/Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue?vue&type=style&index=0&lang=css& ***!
  \*************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

// extracted by mini-css-extract-plugin
    if(false) { var cssReload; }
  

/***/ }),

/***/ 95:
/*!*******************************************************************************************************!*\
  !*** /Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/main.js?{"page":"pages%2Fquiz%2Fquiz"} ***!
  \*******************************************************************************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

"use strict";
/* WEBPACK VAR INJECTION */(function(wx, createPage) {

var _interopRequireDefault = __webpack_require__(/*! @babel/runtime/helpers/interopRequireDefault */ 4);
__webpack_require__(/*! uni-pages */ 26);
var _vue = _interopRequireDefault(__webpack_require__(/*! vue */ 25));
var _quiz = _interopRequireDefault(__webpack_require__(/*! ./pages/quiz/quiz.vue */ 96));
// @ts-ignore
wx.__webpack_require_UNI_MP_PLUGIN__ = __webpack_require__;
createPage(_quiz.default);
/* WEBPACK VAR INJECTION */}.call(this, __webpack_require__(/*! ./node_modules/@dcloudio/uni-mp-weixin/dist/wx.js */ 1)["default"], __webpack_require__(/*! ./node_modules/@dcloudio/uni-mp-weixin/dist/index.js */ 2)["createPage"]))

/***/ }),

/***/ 96:
/*!************************************************************************************!*\
  !*** /Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue ***!
  \************************************************************************************/
/*! no static exports found */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./quiz.vue?vue&type=template&id=47015bb4& */ 97);
/* harmony import */ var _quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ./quiz.vue?vue&type=script&lang=js& */ 99);
/* harmony reexport (unknown) */ for(var __WEBPACK_IMPORT_KEY__ in _quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_1__) if(["default"].indexOf(__WEBPACK_IMPORT_KEY__) < 0) (function(key) { __webpack_require__.d(__webpack_exports__, key, function() { return _quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_1__[key]; }) }(__WEBPACK_IMPORT_KEY__));
/* harmony import */ var _quiz_vue_vue_type_style_index_0_lang_css___WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ./quiz.vue?vue&type=style&index=0&lang=css& */ 101);
/* harmony import */ var _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib/runtime/componentNormalizer.js */ 32);

var renderjs





/* normalize component */

var component = Object(_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_runtime_componentNormalizer_js__WEBPACK_IMPORTED_MODULE_3__["default"])(
  _quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_1__["default"],
  _quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["render"],
  _quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["staticRenderFns"],
  false,
  null,
  null,
  null,
  false,
  _quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["components"],
  renderjs
)

component.options.__file = "pages/quiz/quiz.vue"
/* harmony default export */ __webpack_exports__["default"] = (component.exports);

/***/ }),

/***/ 97:
/*!*******************************************************************************************************************!*\
  !*** /Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue?vue&type=template&id=47015bb4& ***!
  \*******************************************************************************************************************/
/*! exports provided: render, staticRenderFns, recyclableRender, components */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_17_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_template_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_uni_app_loader_page_meta_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-preprocess-loader??ref--17-0!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/webpack-uni-mp-loader/lib/template.js!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-uni-app-loader/page-meta.js!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib??vue-loader-options!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/webpack-uni-mp-loader/lib/style.js!./quiz.vue?vue&type=template&id=47015bb4& */ 98);
/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, "render", function() { return _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_17_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_template_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_uni_app_loader_page_meta_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["render"]; });

/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, "staticRenderFns", function() { return _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_17_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_template_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_uni_app_loader_page_meta_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["staticRenderFns"]; });

/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, "recyclableRender", function() { return _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_17_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_template_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_uni_app_loader_page_meta_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["recyclableRender"]; });

/* harmony reexport (safe) */ __webpack_require__.d(__webpack_exports__, "components", function() { return _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_loaders_templateLoader_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_17_0_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_template_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_uni_app_loader_page_meta_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_template_id_47015bb4___WEBPACK_IMPORTED_MODULE_0__["components"]; });



/***/ }),

/***/ 98:
/*!*******************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************!*\
  !*** ./node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib/loaders/templateLoader.js??vue-loader-options!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-preprocess-loader??ref--17-0!./node_modules/@dcloudio/webpack-uni-mp-loader/lib/template.js!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-uni-app-loader/page-meta.js!./node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib??vue-loader-options!./node_modules/@dcloudio/webpack-uni-mp-loader/lib/style.js!/Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue?vue&type=template&id=47015bb4& ***!
  \*******************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************************/
/*! exports provided: render, staticRenderFns, recyclableRender, components */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "render", function() { return render; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "staticRenderFns", function() { return staticRenderFns; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "recyclableRender", function() { return recyclableRender; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "components", function() { return components; });
var components
var render = function () {
  var _vm = this
  var _h = _vm.$createElement
  var _c = _vm._self._c || _h
  var l0 = !_vm.loading
    ? _vm.__map(_vm.leadInQuestionContent.options, function (option, index) {
        var $orig = _vm.__get_orig(option)
        var m0 = _vm.leadInShowResult && _vm.isLeadInOptionCorrect(index)
        var m1 =
          _vm.leadInShowResult &&
          _vm.leadInSelectedOption === index &&
          !_vm.isLeadInOptionCorrect(index)
        return {
          $orig: $orig,
          m0: m0,
          m1: m1,
        }
      })
    : null
  var l1 =
    !_vm.loading && _vm.showConceptQuestion
      ? _vm.__map(_vm.conceptQuestionContent.options, function (option, index) {
          var $orig = _vm.__get_orig(option)
          var m2 = _vm.conceptShowResult && _vm.isConceptOptionCorrect(index)
          var m3 =
            _vm.conceptShowResult &&
            _vm.conceptSelectedOption === index &&
            !_vm.isConceptOptionCorrect(index)
          return {
            $orig: $orig,
            m2: m2,
            m3: m3,
          }
        })
      : null
  _vm.$mp.data = Object.assign(
    {},
    {
      $root: {
        l0: l0,
        l1: l1,
      },
    }
  )
}
var recyclableRender = false
var staticRenderFns = []
render._withStripped = true



/***/ }),

/***/ 99:
/*!*************************************************************************************************************!*\
  !*** /Users/nuclearwaste/Documents/HBuilderProjects/paperPlay/pages/quiz/quiz.vue?vue&type=script&lang=js& ***!
  \*************************************************************************************************************/
/*! no static exports found */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_babel_loader_lib_index_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_13_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_script_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! -!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/babel-loader/lib!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/webpack-preprocess-loader??ref--13-1!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/webpack-uni-mp-loader/lib/script.js!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/vue-cli-plugin-uni/packages/vue-loader/lib??vue-loader-options!../../../../../../../Applications/HBuilderX.app/Contents/HBuilderX/plugins/uniapp-cli/node_modules/@dcloudio/webpack-uni-mp-loader/lib/style.js!./quiz.vue?vue&type=script&lang=js& */ 100);
/* harmony import */ var _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_babel_loader_lib_index_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_13_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_script_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0___default = /*#__PURE__*/__webpack_require__.n(_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_babel_loader_lib_index_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_13_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_script_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__);
/* harmony reexport (unknown) */ for(var __WEBPACK_IMPORT_KEY__ in _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_babel_loader_lib_index_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_13_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_script_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__) if(["default"].indexOf(__WEBPACK_IMPORT_KEY__) < 0) (function(key) { __webpack_require__.d(__webpack_exports__, key, function() { return _Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_babel_loader_lib_index_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_13_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_script_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0__[key]; }) }(__WEBPACK_IMPORT_KEY__));
 /* harmony default export */ __webpack_exports__["default"] = (_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_babel_loader_lib_index_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_webpack_preprocess_loader_index_js_ref_13_1_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_script_js_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_vue_cli_plugin_uni_packages_vue_loader_lib_index_js_vue_loader_options_Applications_HBuilderX_app_Contents_HBuilderX_plugins_uniapp_cli_node_modules_dcloudio_webpack_uni_mp_loader_lib_style_js_quiz_vue_vue_type_script_lang_js___WEBPACK_IMPORTED_MODULE_0___default.a); 

/***/ })

},[[95,"common/runtime","common/vendor"]]]);
//# sourceMappingURL=../../../.sourcemap/mp-weixin/pages/quiz/quiz.js.map