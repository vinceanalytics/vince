!function(){"use strict";var o=window.location,c=window.document,p=c.currentScript,u=p.getAttribute("data-api")||new URL(p.src).origin+"/api/event";function s(t){console.warn("Ignoring Event: "+t)}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return s("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return s("localStorage flag")}catch(t){}var n={};n.n=t,n.u=e&&e.u?e.u:o.href,n.d=p.getAttribute("data-domain"),n.r=c.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props);var r=p.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)}),i=n.p||{};r.forEach(function(t){var e=t.replace("event-",""),n=p.getAttribute(t);i[e]=i[e]||n}),n.p=i;var a=new XMLHttpRequest;a.open("POST",u,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&e&&e.callback&&e.callback()}}}var e=window.vince&&window.vince.q||[];window.vince=t;for(var n=0;n<e.length;n++)t.apply(this,e[n]);var i=1;function r(t){if("auxclick"!==t.type||t.button===i){var e,n=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target),r=n&&n.href&&n.href.split("?")[0];return(e=n)&&e.href&&e.host&&e.host!==o.host?a(t,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(t){if(!t)return!1;var e=t.split(".").pop();return v.some(function(t){return t===e})}(r)?a(t,n,{name:"File Download",props:{url:r}}):void 0}}function a(t,e,n){var r=!1;function i(){r||(r=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),r=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&r}}(t,e)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:i}),setTimeout(i,5e3),t.preventDefault())}c.addEventListener("click",r),c.addEventListener("auxclick",r);var l=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],f=p.getAttribute("file-types"),d=p.getAttribute("add-file-types"),v=f&&f.split(",")||d&&d.split(",").concat(l)||l}();