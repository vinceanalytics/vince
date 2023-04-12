!function(){"use strict";var s=window.location,f=window.document,d=f.currentScript,v=d.getAttribute("data-api")||new URL(d.src).origin+"/api/event";function w(t){console.warn("Ignoring Event: "+t)}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return w("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return w("localStorage flag")}catch(t){}var r=d&&d.getAttribute("data-include"),n=d&&d.getAttribute("data-exclude");if("pageview"===t){var i=!r||r&&r.split(",").some(l),a=n&&n.split(",").some(l);if(!i||a)return w("exclusion rule")}var o={};o.n=t,o.u=e&&e.u?e.u:s.href,o.d=d.getAttribute("data-domain"),o.r=f.referrer||null,o.w=window.innerWidth,e&&e.meta&&(o.m=JSON.stringify(e.meta)),e&&e.props&&(o.p=e.props);var c=d.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)}),p=o.p||{};c.forEach(function(t){var e=t.replace("event-",""),r=d.getAttribute(t);p[e]=p[e]||r}),o.p=p,o.h=1;var u=new XMLHttpRequest;u.open("POST",v,!0),u.setRequestHeader("Content-Type","text/plain"),u.send(JSON.stringify(o)),u.onreadystatechange=function(){4===u.readyState&&e&&e.callback&&e.callback()}}function l(t){var e=s.pathname;return(e+=s.hash).match(new RegExp("^"+t.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var e=window.vince&&window.vince.q||[];window.vince=t;for(var r=0;r<e.length;r++)t.apply(this,e[r]);var i=1;function n(t){if("auxclick"!==t.type||t.button===i){var e,r=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target),n=r&&r.href&&r.href.split("?")[0];return(e=r)&&e.href&&e.host&&e.host!==s.host?a(t,r,{name:"Outbound Link: Click",props:{url:r.href}}):function(t){if(!t)return!1;var e=t.split(".").pop();return u.some(function(t){return t===e})}(n)?a(t,r,{name:"File Download",props:{url:n}}):void 0}}function a(t,e,r){var n=!1;function i(){n||(n=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented){var r=!e.target||e.target.match(/^_(self|parent|top)$/i),n=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return r&&n}}(t,e)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:i}),setTimeout(i,5e3),t.preventDefault())}f.addEventListener("click",n),f.addEventListener("auxclick",n);var o=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],c=d.getAttribute("file-types"),p=d.getAttribute("add-file-types"),u=c&&c.split(",")||p&&p.split(",").concat(o)||o}();