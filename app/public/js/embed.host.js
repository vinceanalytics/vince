(() => {
  var __create = Object.create;
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __getProtoOf = Object.getPrototypeOf;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __commonJS = (cb, mod) => function __require() {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
    // If the importer is in node compatibility mode or this is not an ESM
    // file that has been converted to a CommonJS file using a Babel-
    // compatible transform (i.e. "__esModule" has not been set), then set
    // "default" to the CommonJS "module.exports" for node compatibility.
    isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
    mod
  ));

  // node_modules/iframe-resizer/js/iframeResizer.js
  var require_iframeResizer = __commonJS({
    "node_modules/iframe-resizer/js/iframeResizer.js"(exports, module) {
      (function(undefined) {
        if (typeof window === "undefined") return;
        var count = 0, logEnabled = false, hiddenCheckEnabled = false, msgHeader = "message", msgHeaderLen = msgHeader.length, msgId = "[iFrameSizer]", msgIdLen = msgId.length, pagePosition = null, requestAnimationFrame = window.requestAnimationFrame, resetRequiredMethods = {
          max: 1,
          scroll: 1,
          bodyScroll: 1,
          documentElementScroll: 1
        }, settings = {}, timer = null, defaults = {
          autoResize: true,
          bodyBackground: null,
          bodyMargin: null,
          bodyMarginV1: 8,
          bodyPadding: null,
          checkOrigin: true,
          inPageLinks: false,
          enablePublicMethods: true,
          heightCalculationMethod: "bodyOffset",
          id: "iFrameResizer",
          interval: 32,
          log: false,
          maxHeight: Infinity,
          maxWidth: Infinity,
          minHeight: 0,
          minWidth: 0,
          mouseEvents: true,
          resizeFrom: "parent",
          scrolling: false,
          sizeHeight: true,
          sizeWidth: false,
          warningTimeout: 5e3,
          tolerance: 0,
          widthCalculationMethod: "scroll",
          onClose: function() {
            return true;
          },
          onClosed: function() {
          },
          onInit: function() {
          },
          onMessage: function() {
            warn("onMessage function not defined");
          },
          onMouseEnter: function() {
          },
          onMouseLeave: function() {
          },
          onResized: function() {
          },
          onScroll: function() {
            return true;
          }
        };
        function getMutationObserver() {
          return window.MutationObserver || window.WebKitMutationObserver || window.MozMutationObserver;
        }
        function addEventListener(el, evt, func) {
          el.addEventListener(evt, func, false);
        }
        function removeEventListener(el, evt, func) {
          el.removeEventListener(evt, func, false);
        }
        function setupRequestAnimationFrame() {
          var vendors = ["moz", "webkit", "o", "ms"];
          var x;
          for (x = 0; x < vendors.length && !requestAnimationFrame; x += 1) {
            requestAnimationFrame = window[vendors[x] + "RequestAnimationFrame"];
          }
          if (!requestAnimationFrame) {
            log("setup", "RequestAnimationFrame not supported");
          } else {
            requestAnimationFrame = requestAnimationFrame.bind(window);
          }
        }
        function getMyID(iframeId) {
          var retStr = "Host page: " + iframeId;
          if (window.top !== window.self) {
            retStr = window.parentIFrame && window.parentIFrame.getId ? window.parentIFrame.getId() + ": " + iframeId : "Nested host page: " + iframeId;
          }
          return retStr;
        }
        function formatLogHeader(iframeId) {
          return msgId + "[" + getMyID(iframeId) + "]";
        }
        function isLogEnabled(iframeId) {
          return settings[iframeId] ? settings[iframeId].log : logEnabled;
        }
        function log(iframeId, msg) {
          output("log", iframeId, msg, isLogEnabled(iframeId));
        }
        function info(iframeId, msg) {
          output("info", iframeId, msg, isLogEnabled(iframeId));
        }
        function warn(iframeId, msg) {
          output("warn", iframeId, msg, true);
        }
        function output(type, iframeId, msg, enabled) {
          if (true === enabled && "object" === typeof window.console) {
            console[type](formatLogHeader(iframeId), msg);
          }
        }
        function iFrameListener(event) {
          function resizeIFrame() {
            function resize() {
              setSize(messageData);
              setPagePosition(iframeId);
              on("onResized", messageData);
            }
            ensureInRange("Height");
            ensureInRange("Width");
            syncResize(resize, messageData, "init");
          }
          function processMsg() {
            var data = msg.substr(msgIdLen).split(":");
            var height = data[1] ? parseInt(data[1], 10) : 0;
            var iframe = settings[data[0]] && settings[data[0]].iframe;
            var compStyle = getComputedStyle(iframe);
            return {
              iframe,
              id: data[0],
              height: height + getPaddingEnds(compStyle) + getBorderEnds(compStyle),
              width: data[2],
              type: data[3]
            };
          }
          function getPaddingEnds(compStyle) {
            if (compStyle.boxSizing !== "border-box") {
              return 0;
            }
            var top = compStyle.paddingTop ? parseInt(compStyle.paddingTop, 10) : 0;
            var bot = compStyle.paddingBottom ? parseInt(compStyle.paddingBottom, 10) : 0;
            return top + bot;
          }
          function getBorderEnds(compStyle) {
            if (compStyle.boxSizing !== "border-box") {
              return 0;
            }
            var top = compStyle.borderTopWidth ? parseInt(compStyle.borderTopWidth, 10) : 0;
            var bot = compStyle.borderBottomWidth ? parseInt(compStyle.borderBottomWidth, 10) : 0;
            return top + bot;
          }
          function ensureInRange(Dimension) {
            var max = Number(settings[iframeId]["max" + Dimension]), min = Number(settings[iframeId]["min" + Dimension]), dimension = Dimension.toLowerCase(), size = Number(messageData[dimension]);
            log(iframeId, "Checking " + dimension + " is in range " + min + "-" + max);
            if (size < min) {
              size = min;
              log(iframeId, "Set " + dimension + " to min value");
            }
            if (size > max) {
              size = max;
              log(iframeId, "Set " + dimension + " to max value");
            }
            messageData[dimension] = "" + size;
          }
          function isMessageFromIFrame() {
            function checkAllowedOrigin() {
              function checkList() {
                var i = 0, retCode = false;
                log(
                  iframeId,
                  "Checking connection is from allowed list of origins: " + checkOrigin
                );
                for (; i < checkOrigin.length; i++) {
                  if (checkOrigin[i] === origin) {
                    retCode = true;
                    break;
                  }
                }
                return retCode;
              }
              function checkSingle() {
                var remoteHost = settings[iframeId] && settings[iframeId].remoteHost;
                log(iframeId, "Checking connection is from: " + remoteHost);
                return origin === remoteHost;
              }
              return checkOrigin.constructor === Array ? checkList() : checkSingle();
            }
            var origin = event.origin, checkOrigin = settings[iframeId] && settings[iframeId].checkOrigin;
            if (checkOrigin && "" + origin !== "null" && !checkAllowedOrigin()) {
              throw new Error(
                "Unexpected message received from: " + origin + " for " + messageData.iframe.id + ". Message was: " + event.data + ". This error can be disabled by setting the checkOrigin: false option or by providing of array of trusted domains."
              );
            }
            return true;
          }
          function isMessageForUs() {
            return msgId === ("" + msg).substr(0, msgIdLen) && msg.substr(msgIdLen).split(":")[0] in settings;
          }
          function isMessageFromMetaParent() {
            var retCode = messageData.type in { true: 1, false: 1, undefined: 1 };
            if (retCode) {
              log(iframeId, "Ignoring init message from meta parent page");
            }
            return retCode;
          }
          function getMsgBody(offset) {
            return msg.substr(msg.indexOf(":") + msgHeaderLen + offset);
          }
          function forwardMsgFromIFrame(msgBody) {
            log(
              iframeId,
              "onMessage passed: {iframe: " + messageData.iframe.id + ", message: " + msgBody + "}"
            );
            on("onMessage", {
              iframe: messageData.iframe,
              message: JSON.parse(msgBody)
            });
            log(iframeId, "--");
          }
          function getPageInfo() {
            var bodyPosition = document.body.getBoundingClientRect(), iFramePosition = messageData.iframe.getBoundingClientRect();
            return JSON.stringify({
              iframeHeight: iFramePosition.height,
              iframeWidth: iFramePosition.width,
              clientHeight: Math.max(
                document.documentElement.clientHeight,
                window.innerHeight || 0
              ),
              clientWidth: Math.max(
                document.documentElement.clientWidth,
                window.innerWidth || 0
              ),
              offsetTop: parseInt(iFramePosition.top - bodyPosition.top, 10),
              offsetLeft: parseInt(iFramePosition.left - bodyPosition.left, 10),
              scrollTop: window.pageYOffset,
              scrollLeft: window.pageXOffset,
              documentHeight: document.documentElement.clientHeight,
              documentWidth: document.documentElement.clientWidth,
              windowHeight: window.innerHeight,
              windowWidth: window.innerWidth
            });
          }
          function sendPageInfoToIframe(iframe, iframeId2) {
            function debouncedTrigger() {
              trigger("Send Page Info", "pageInfo:" + getPageInfo(), iframe, iframeId2);
            }
            debounceFrameEvents(debouncedTrigger, 32, iframeId2);
          }
          function startPageInfoMonitor() {
            function setListener(type, func) {
              function sendPageInfo() {
                if (settings[id]) {
                  sendPageInfoToIframe(settings[id].iframe, id);
                } else {
                  stop();
                }
              }
              ;
              ["scroll", "resize"].forEach(function(evt) {
                log(id, type + evt + " listener for sendPageInfo");
                func(window, evt, sendPageInfo);
              });
            }
            function stop() {
              setListener("Remove ", removeEventListener);
            }
            function start() {
              setListener("Add ", addEventListener);
            }
            var id = iframeId;
            start();
            if (settings[id]) {
              settings[id].stopPageInfo = stop;
            }
          }
          function stopPageInfoMonitor() {
            if (settings[iframeId] && settings[iframeId].stopPageInfo) {
              settings[iframeId].stopPageInfo();
              delete settings[iframeId].stopPageInfo;
            }
          }
          function checkIFrameExists() {
            var retBool = true;
            if (null === messageData.iframe) {
              warn(iframeId, "IFrame (" + messageData.id + ") not found");
              retBool = false;
            }
            return retBool;
          }
          function getElementPosition(target) {
            var iFramePosition = target.getBoundingClientRect();
            getPagePosition(iframeId);
            return {
              x: Math.floor(Number(iFramePosition.left) + Number(pagePosition.x)),
              y: Math.floor(Number(iFramePosition.top) + Number(pagePosition.y))
            };
          }
          function scrollRequestFromChild(addOffset) {
            function reposition() {
              pagePosition = newPosition;
              scrollTo();
              log(iframeId, "--");
            }
            function calcOffset() {
              return {
                x: Number(messageData.width) + offset.x,
                y: Number(messageData.height) + offset.y
              };
            }
            function scrollParent() {
              if (window.parentIFrame) {
                window.parentIFrame["scrollTo" + (addOffset ? "Offset" : "")](
                  newPosition.x,
                  newPosition.y
                );
              } else {
                warn(
                  iframeId,
                  "Unable to scroll to requested position, window.parentIFrame not found"
                );
              }
            }
            var offset = addOffset ? getElementPosition(messageData.iframe) : { x: 0, y: 0 }, newPosition = calcOffset();
            log(
              iframeId,
              "Reposition requested from iFrame (offset x:" + offset.x + " y:" + offset.y + ")"
            );
            if (window.top !== window.self) {
              scrollParent();
            } else {
              reposition();
            }
          }
          function scrollTo() {
            if (false !== on("onScroll", pagePosition)) {
              setPagePosition(iframeId);
            } else {
              unsetPagePosition();
            }
          }
          function findTarget(location) {
            function jumpToTarget() {
              var jumpPosition = getElementPosition(target);
              log(
                iframeId,
                "Moving to in page link (#" + hash + ") at x: " + jumpPosition.x + " y: " + jumpPosition.y
              );
              pagePosition = {
                x: jumpPosition.x,
                y: jumpPosition.y
              };
              scrollTo();
              log(iframeId, "--");
            }
            function jumpToParent() {
              if (window.parentIFrame) {
                window.parentIFrame.moveToAnchor(hash);
              } else {
                log(
                  iframeId,
                  "In page link #" + hash + " not found and window.parentIFrame not found"
                );
              }
            }
            var hash = location.split("#")[1] || "", hashData = decodeURIComponent(hash), target = document.getElementById(hashData) || document.getElementsByName(hashData)[0];
            if (target) {
              jumpToTarget();
            } else if (window.top !== window.self) {
              jumpToParent();
            } else {
              log(iframeId, "In page link #" + hash + " not found");
            }
          }
          function onMouse(event2) {
            var mousePos = {};
            if (Number(messageData.width) === 0 && Number(messageData.height) === 0) {
              var data = getMsgBody(9).split(":");
              mousePos = {
                x: data[1],
                y: data[0]
              };
            } else {
              mousePos = {
                x: messageData.width,
                y: messageData.height
              };
            }
            on(event2, {
              iframe: messageData.iframe,
              screenX: Number(mousePos.x),
              screenY: Number(mousePos.y),
              type: messageData.type
            });
          }
          function on(funcName, val) {
            return chkEvent(iframeId, funcName, val);
          }
          function actionMsg() {
            if (settings[iframeId] && settings[iframeId].firstRun) firstRun();
            switch (messageData.type) {
              case "close":
                closeIFrame(messageData.iframe);
                break;
              case "message":
                forwardMsgFromIFrame(getMsgBody(6));
                break;
              case "mouseenter":
                onMouse("onMouseEnter");
                break;
              case "mouseleave":
                onMouse("onMouseLeave");
                break;
              case "autoResize":
                settings[iframeId].autoResize = JSON.parse(getMsgBody(9));
                break;
              case "scrollTo":
                scrollRequestFromChild(false);
                break;
              case "scrollToOffset":
                scrollRequestFromChild(true);
                break;
              case "pageInfo":
                sendPageInfoToIframe(
                  settings[iframeId] && settings[iframeId].iframe,
                  iframeId
                );
                startPageInfoMonitor();
                break;
              case "pageInfoStop":
                stopPageInfoMonitor();
                break;
              case "inPageLink":
                findTarget(getMsgBody(9));
                break;
              case "reset":
                resetIFrame(messageData);
                break;
              case "init":
                resizeIFrame();
                on("onInit", messageData.iframe);
                break;
              default:
                if (Number(messageData.width) === 0 && Number(messageData.height) === 0) {
                  warn(
                    "Unsupported message received (" + messageData.type + "), this is likely due to the iframe containing a later version of iframe-resizer than the parent page"
                  );
                } else {
                  resizeIFrame();
                }
            }
          }
          function hasSettings(iframeId2) {
            var retBool = true;
            if (!settings[iframeId2]) {
              retBool = false;
              warn(
                messageData.type + " No settings for " + iframeId2 + ". Message was: " + msg
              );
            }
            return retBool;
          }
          function iFrameReadyMsgReceived() {
            for (var iframeId2 in settings) {
              trigger(
                "iFrame requested init",
                createOutgoingMsg(iframeId2),
                settings[iframeId2].iframe,
                iframeId2
              );
            }
          }
          function firstRun() {
            if (settings[iframeId]) {
              settings[iframeId].firstRun = false;
            }
          }
          var msg = event.data, messageData = {}, iframeId = null;
          if ("[iFrameResizerChild]Ready" === msg) {
            iFrameReadyMsgReceived();
          } else if (isMessageForUs()) {
            messageData = processMsg();
            iframeId = messageData.id;
            if (settings[iframeId]) {
              settings[iframeId].loaded = true;
            }
            if (!isMessageFromMetaParent() && hasSettings(iframeId)) {
              log(iframeId, "Received: " + msg);
              if (checkIFrameExists() && isMessageFromIFrame()) {
                actionMsg();
              }
            }
          } else {
            info(iframeId, "Ignored: " + msg);
          }
        }
        function chkEvent(iframeId, funcName, val) {
          var func = null, retVal = null;
          if (settings[iframeId]) {
            func = settings[iframeId][funcName];
            if ("function" === typeof func) {
              retVal = func(val);
            } else {
              throw new TypeError(
                funcName + " on iFrame[" + iframeId + "] is not a function"
              );
            }
          }
          return retVal;
        }
        function removeIframeListeners(iframe) {
          var iframeId = iframe.id;
          delete settings[iframeId];
        }
        function closeIFrame(iframe) {
          var iframeId = iframe.id;
          if (chkEvent(iframeId, "onClose", iframeId) === false) {
            log(iframeId, "Close iframe cancelled by onClose event");
            return;
          }
          log(iframeId, "Removing iFrame: " + iframeId);
          try {
            if (iframe.parentNode) {
              iframe.parentNode.removeChild(iframe);
            }
          } catch (error) {
            warn(error);
          }
          chkEvent(iframeId, "onClosed", iframeId);
          log(iframeId, "--");
          removeIframeListeners(iframe);
        }
        function getPagePosition(iframeId) {
          if (null === pagePosition) {
            pagePosition = {
              x: window.pageXOffset !== undefined ? window.pageXOffset : document.documentElement.scrollLeft,
              y: window.pageYOffset !== undefined ? window.pageYOffset : document.documentElement.scrollTop
            };
            log(
              iframeId,
              "Get page position: " + pagePosition.x + "," + pagePosition.y
            );
          }
        }
        function setPagePosition(iframeId) {
          if (null !== pagePosition) {
            window.scrollTo(pagePosition.x, pagePosition.y);
            log(
              iframeId,
              "Set page position: " + pagePosition.x + "," + pagePosition.y
            );
            unsetPagePosition();
          }
        }
        function unsetPagePosition() {
          pagePosition = null;
        }
        function resetIFrame(messageData) {
          function reset() {
            setSize(messageData);
            trigger("reset", "reset", messageData.iframe, messageData.id);
          }
          log(
            messageData.id,
            "Size reset requested by " + ("init" === messageData.type ? "host page" : "iFrame")
          );
          getPagePosition(messageData.id);
          syncResize(reset, messageData, "reset");
        }
        function setSize(messageData) {
          function setDimension(dimension) {
            if (!messageData.id) {
              log("undefined", "messageData id not set");
              return;
            }
            messageData.iframe.style[dimension] = messageData[dimension] + "px";
            log(
              messageData.id,
              "IFrame (" + iframeId + ") " + dimension + " set to " + messageData[dimension] + "px"
            );
          }
          function chkZero(dimension) {
            if (!hiddenCheckEnabled && "0" === messageData[dimension]) {
              hiddenCheckEnabled = true;
              log(iframeId, "Hidden iFrame detected, creating visibility listener");
              fixHiddenIFrames();
            }
          }
          function processDimension(dimension) {
            setDimension(dimension);
            chkZero(dimension);
          }
          var iframeId = messageData.iframe.id;
          if (settings[iframeId]) {
            if (settings[iframeId].sizeHeight) {
              processDimension("height");
            }
            if (settings[iframeId].sizeWidth) {
              processDimension("width");
            }
          }
        }
        function syncResize(func, messageData, doNotSync) {
          if (doNotSync !== messageData.type && requestAnimationFrame && // including check for jasmine because had trouble getting spy to work in unit test using requestAnimationFrame
          !window.jasmine) {
            log(messageData.id, "Requesting animation frame");
            requestAnimationFrame(func);
          } else {
            func();
          }
        }
        function trigger(calleeMsg, msg, iframe, id, noResponseWarning) {
          function postMessageToIFrame() {
            var target = settings[id] && settings[id].targetOrigin;
            log(
              id,
              "[" + calleeMsg + "] Sending msg to iframe[" + id + "] (" + msg + ") targetOrigin: " + target
            );
            iframe.contentWindow.postMessage(msgId + msg, target);
          }
          function iFrameNotFound() {
            warn(id, "[" + calleeMsg + "] IFrame(" + id + ") not found");
          }
          function chkAndSend() {
            if (iframe && "contentWindow" in iframe && null !== iframe.contentWindow) {
              postMessageToIFrame();
            } else {
              iFrameNotFound();
            }
          }
          function warnOnNoResponse() {
            function warning() {
              if (settings[id] && !settings[id].loaded && !errorShown) {
                errorShown = true;
                warn(
                  id,
                  "IFrame has not responded within " + settings[id].warningTimeout / 1e3 + " seconds. Check iFrameResizer.contentWindow.js has been loaded in iFrame. This message can be ignored if everything is working, or you can set the warningTimeout option to a higher value or zero to suppress this warning."
                );
              }
            }
            if (!!noResponseWarning && settings[id] && !!settings[id].warningTimeout) {
              settings[id].msgTimeout = setTimeout(
                warning,
                settings[id].warningTimeout
              );
            }
          }
          var errorShown = false;
          id = id || iframe.id;
          if (settings[id]) {
            chkAndSend();
            warnOnNoResponse();
          }
        }
        function createOutgoingMsg(iframeId) {
          return iframeId + ":" + settings[iframeId].bodyMarginV1 + ":" + settings[iframeId].sizeWidth + ":" + settings[iframeId].log + ":" + settings[iframeId].interval + ":" + settings[iframeId].enablePublicMethods + ":" + settings[iframeId].autoResize + ":" + settings[iframeId].bodyMargin + ":" + settings[iframeId].heightCalculationMethod + ":" + settings[iframeId].bodyBackground + ":" + settings[iframeId].bodyPadding + ":" + settings[iframeId].tolerance + ":" + settings[iframeId].inPageLinks + ":" + settings[iframeId].resizeFrom + ":" + settings[iframeId].widthCalculationMethod + ":" + settings[iframeId].mouseEvents;
        }
        function isNumber(value) {
          return typeof value === "number";
        }
        function setupIFrame(iframe, options) {
          function setLimits() {
            function addStyle(style) {
              var styleValue = settings[iframeId][style];
              if (Infinity !== styleValue && 0 !== styleValue) {
                iframe.style[style] = isNumber(styleValue) ? styleValue + "px" : styleValue;
                log(iframeId, "Set " + style + " = " + iframe.style[style]);
              }
            }
            function chkMinMax(dimension) {
              if (settings[iframeId]["min" + dimension] > settings[iframeId]["max" + dimension]) {
                throw new Error(
                  "Value for min" + dimension + " can not be greater than max" + dimension
                );
              }
            }
            chkMinMax("Height");
            chkMinMax("Width");
            addStyle("maxHeight");
            addStyle("minHeight");
            addStyle("maxWidth");
            addStyle("minWidth");
          }
          function newId() {
            var id = options && options.id || defaults.id + count++;
            if (null !== document.getElementById(id)) {
              id += count++;
            }
            return id;
          }
          function ensureHasId(iframeId2) {
            if ("" === iframeId2) {
              iframe.id = iframeId2 = newId();
              logEnabled = (options || {}).log;
              log(
                iframeId2,
                "Added missing iframe ID: " + iframeId2 + " (" + iframe.src + ")"
              );
            }
            return iframeId2;
          }
          function setScrolling() {
            log(
              iframeId,
              "IFrame scrolling " + (settings[iframeId] && settings[iframeId].scrolling ? "enabled" : "disabled") + " for " + iframeId
            );
            iframe.style.overflow = false === (settings[iframeId] && settings[iframeId].scrolling) ? "hidden" : "auto";
            switch (settings[iframeId] && settings[iframeId].scrolling) {
              case "omit":
                break;
              case true:
                iframe.scrolling = "yes";
                break;
              case false:
                iframe.scrolling = "no";
                break;
              default:
                iframe.scrolling = settings[iframeId] ? settings[iframeId].scrolling : "no";
            }
          }
          function setupBodyMarginValues() {
            if ("number" === typeof (settings[iframeId] && settings[iframeId].bodyMargin) || "0" === (settings[iframeId] && settings[iframeId].bodyMargin)) {
              settings[iframeId].bodyMarginV1 = settings[iframeId].bodyMargin;
              settings[iframeId].bodyMargin = "" + settings[iframeId].bodyMargin + "px";
            }
          }
          function checkReset() {
            var firstRun = settings[iframeId] && settings[iframeId].firstRun, resetRequertMethod = settings[iframeId] && settings[iframeId].heightCalculationMethod in resetRequiredMethods;
            if (!firstRun && resetRequertMethod) {
              resetIFrame({ iframe, height: 0, width: 0, type: "init" });
            }
          }
          function setupIFrameObject() {
            if (settings[iframeId]) {
              settings[iframeId].iframe.iFrameResizer = {
                close: closeIFrame.bind(null, settings[iframeId].iframe),
                removeListeners: removeIframeListeners.bind(
                  null,
                  settings[iframeId].iframe
                ),
                resize: trigger.bind(
                  null,
                  "Window resize",
                  "resize",
                  settings[iframeId].iframe
                ),
                moveToAnchor: function(anchor) {
                  trigger(
                    "Move to anchor",
                    "moveToAnchor:" + anchor,
                    settings[iframeId].iframe,
                    iframeId
                  );
                },
                sendMessage: function(message) {
                  message = JSON.stringify(message);
                  trigger(
                    "Send Message",
                    "message:" + message,
                    settings[iframeId].iframe,
                    iframeId
                  );
                }
              };
            }
          }
          function init(msg) {
            function iFrameLoaded() {
              trigger("iFrame.onload", msg, iframe, undefined, true);
              checkReset();
            }
            function createDestroyObserver(MutationObserver2) {
              if (!iframe.parentNode) {
                return;
              }
              var destroyObserver = new MutationObserver2(function(mutations) {
                mutations.forEach(function(mutation) {
                  var removedNodes = Array.prototype.slice.call(mutation.removedNodes);
                  removedNodes.forEach(function(removedNode) {
                    if (removedNode === iframe) {
                      closeIFrame(iframe);
                    }
                  });
                });
              });
              destroyObserver.observe(iframe.parentNode, {
                childList: true
              });
            }
            var MutationObserver = getMutationObserver();
            if (MutationObserver) {
              createDestroyObserver(MutationObserver);
            }
            addEventListener(iframe, "load", iFrameLoaded);
            trigger("init", msg, iframe, undefined, true);
          }
          function checkOptions(options2) {
            if ("object" !== typeof options2) {
              throw new TypeError("Options is not an object");
            }
          }
          function copyOptions(options2) {
            for (var option in defaults) {
              if (Object.prototype.hasOwnProperty.call(defaults, option)) {
                settings[iframeId][option] = Object.prototype.hasOwnProperty.call(
                  options2,
                  option
                ) ? options2[option] : defaults[option];
              }
            }
          }
          function getTargetOrigin(remoteHost) {
            return "" === remoteHost || null !== remoteHost.match(/^(about:blank|javascript:|file:\/\/)/) ? "*" : remoteHost;
          }
          function depricate(key) {
            var splitName = key.split("Callback");
            if (splitName.length === 2) {
              var name = "on" + splitName[0].charAt(0).toUpperCase() + splitName[0].slice(1);
              this[name] = this[key];
              delete this[key];
              warn(
                iframeId,
                "Deprecated: '" + key + "' has been renamed '" + name + "'. The old method will be removed in the next major version."
              );
            }
          }
          function processOptions(options2) {
            options2 = options2 || {};
            settings[iframeId] = {
              firstRun: true,
              iframe,
              remoteHost: iframe.src && iframe.src.split("/").slice(0, 3).join("/")
            };
            checkOptions(options2);
            Object.keys(options2).forEach(depricate, options2);
            copyOptions(options2);
            if (settings[iframeId]) {
              settings[iframeId].targetOrigin = true === settings[iframeId].checkOrigin ? getTargetOrigin(settings[iframeId].remoteHost) : "*";
            }
          }
          function beenHere() {
            return iframeId in settings && "iFrameResizer" in iframe;
          }
          var iframeId = ensureHasId(iframe.id);
          if (!beenHere()) {
            processOptions(options);
            setScrolling();
            setLimits();
            setupBodyMarginValues();
            init(createOutgoingMsg(iframeId));
            setupIFrameObject();
          } else {
            warn(iframeId, "Ignored iFrame, already setup.");
          }
        }
        function debouce(fn, time) {
          if (null === timer) {
            timer = setTimeout(function() {
              timer = null;
              fn();
            }, time);
          }
        }
        var frameTimer = {};
        function debounceFrameEvents(fn, time, frameId) {
          if (!frameTimer[frameId]) {
            frameTimer[frameId] = setTimeout(function() {
              frameTimer[frameId] = null;
              fn();
            }, time);
          }
        }
        function fixHiddenIFrames() {
          function checkIFrames() {
            function checkIFrame(settingId) {
              function chkDimension(dimension) {
                return "0px" === (settings[settingId] && settings[settingId].iframe.style[dimension]);
              }
              function isVisible(el) {
                return null !== el.offsetParent;
              }
              if (settings[settingId] && isVisible(settings[settingId].iframe) && (chkDimension("height") || chkDimension("width"))) {
                trigger(
                  "Visibility change",
                  "resize",
                  settings[settingId].iframe,
                  settingId
                );
              }
            }
            Object.keys(settings).forEach(function(key) {
              checkIFrame(key);
            });
          }
          function mutationObserved(mutations) {
            log(
              "window",
              "Mutation observed: " + mutations[0].target + " " + mutations[0].type
            );
            debouce(checkIFrames, 16);
          }
          function createMutationObserver() {
            var target = document.querySelector("body"), config = {
              attributes: true,
              attributeOldValue: false,
              characterData: true,
              characterDataOldValue: false,
              childList: true,
              subtree: true
            }, observer = new MutationObserver(mutationObserved);
            observer.observe(target, config);
          }
          var MutationObserver = getMutationObserver();
          if (MutationObserver) {
            createMutationObserver();
          }
        }
        function resizeIFrames(event) {
          function resize() {
            sendTriggerMsg("Window " + event, "resize");
          }
          log("window", "Trigger event: " + event);
          debouce(resize, 16);
        }
        function tabVisible() {
          function resize() {
            sendTriggerMsg("Tab Visable", "resize");
          }
          if ("hidden" !== document.visibilityState) {
            log("document", "Trigger event: Visiblity change");
            debouce(resize, 16);
          }
        }
        function sendTriggerMsg(eventName, event) {
          function isIFrameResizeEnabled(iframeId) {
            return settings[iframeId] && "parent" === settings[iframeId].resizeFrom && settings[iframeId].autoResize && !settings[iframeId].firstRun;
          }
          Object.keys(settings).forEach(function(iframeId) {
            if (isIFrameResizeEnabled(iframeId)) {
              trigger(eventName, event, settings[iframeId].iframe, iframeId);
            }
          });
        }
        function setupEventListeners() {
          addEventListener(window, "message", iFrameListener);
          addEventListener(window, "resize", function() {
            resizeIFrames("resize");
          });
          addEventListener(document, "visibilitychange", tabVisible);
          addEventListener(document, "-webkit-visibilitychange", tabVisible);
        }
        function factory() {
          function init(options, element) {
            function chkType() {
              if (!element.tagName) {
                throw new TypeError("Object is not a valid DOM element");
              } else if ("IFRAME" !== element.tagName.toUpperCase()) {
                throw new TypeError(
                  "Expected <IFRAME> tag, found <" + element.tagName + ">"
                );
              }
            }
            if (element) {
              chkType();
              setupIFrame(element, options);
              iFrames.push(element);
            }
          }
          function warnDeprecatedOptions(options) {
            if (options && options.enablePublicMethods) {
              warn(
                "enablePublicMethods option has been removed, public methods are now always available in the iFrame"
              );
            }
          }
          var iFrames;
          setupRequestAnimationFrame();
          setupEventListeners();
          return function iFrameResizeF(options, target) {
            iFrames = [];
            warnDeprecatedOptions(options);
            switch (typeof target) {
              case "undefined":
              case "string":
                Array.prototype.forEach.call(
                  document.querySelectorAll(target || "iframe"),
                  init.bind(undefined, options)
                );
                break;
              case "object":
                init(options, target);
                break;
              default:
                throw new TypeError("Unexpected data type (" + typeof target + ")");
            }
            return iFrames;
          };
        }
        function createJQueryPublicMethod($) {
          if (!$.fn) {
            info("", "Unable to bind to jQuery, it is not fully loaded.");
          } else if (!$.fn.iFrameResize) {
            $.fn.iFrameResize = function $iFrameResizeF(options) {
              function init(index, element) {
                setupIFrame(element, options);
              }
              return this.filter("iframe").each(init).end();
            };
          }
        }
        if (window.jQuery) {
          createJQueryPublicMethod(window.jQuery);
        }
        if (typeof define === "function" && define.amd) {
          define([], factory);
        } else if (typeof module === "object" && typeof module.exports === "object") {
          module.exports = factory();
        }
        window.iFrameResize = window.iFrameResize || factory();
      })();
    }
  });

  // js/embed.host.js
  var import_iframeResizer = __toESM(require_iframeResizer());
  var iframes = (0, import_iframeResizer.default)({
    heightCalculationMethod: "taggedElement",
    onInit,
    checkOrigin: false
  }, "[plausible-embed]");
  function onInit() {
    var iframe = iframes[0];
    var styles = iframe.getAttribute("styles");
    if (styles) {
      iframe.iFrameResizer.sendMessage({
        type: "load-custom-styles",
        opts: {
          styles
        }
      });
    }
  }
})();
