!function(){"use strict";var a=window.location,r=window.document,o=r.currentScript,p=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function c(t){console.warn("Ignoring Event: "+t)}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(a.hostname)||"file:"===a.protocol)return c("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return c("localStorage flag")}catch(t){}var n={};n.n=t,n.u=a.href,n.d=o.getAttribute("data-domain"),n.r=r.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props);var i=new XMLHttpRequest;i.open("POST",p,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&e&&e.callback&&e.callback()}}}var e=window.vince&&window.vince.q||[];window.vince=t;for(var n,i=0;i<e.length;i++)t.apply(this,e[i]);function s(){n!==a.pathname&&(n=a.pathname,t("pageview"))}var l,d=window.history;d.pushState&&(l=d.pushState,d.pushState=function(){l.apply(this,arguments),s()},window.addEventListener("popstate",s)),"prerender"===r.visibilityState?r.addEventListener("visibilitychange",function(){n||"visible"!==r.visibilityState||s()}):s();var u=1;function f(t){if("auxclick"!==t.type||t.button===u){var e,n,i,a,r=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target),o=r&&r.href&&r.href.split("?")[0];if(function(t){if(!t)return!1;var e=t.split(".").pop();return m.some(function(t){return t===e})}(o))return a=!(i={name:"File Download",props:{url:o}}),void(!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),i=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&i}}(e=t,n=r)?vince(i.name,{props:i.props}):(vince(i.name,{props:i.props,callback:p}),setTimeout(p,5e3),e.preventDefault()))}function p(){a||(a=!0,window.location=n.href)}}r.addEventListener("click",f),r.addEventListener("auxclick",f);var v=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],w=o.getAttribute("file-types"),g=o.getAttribute("add-file-types"),m=w&&w.split(",")||g&&g.split(",").concat(v)||v}();