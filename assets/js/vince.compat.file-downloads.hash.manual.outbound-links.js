!function(){"use strict";var t,e,n,i=window.location,a=window.document,o=a.getElementById("vince"),p=o.getAttribute("data-api")||(t=o.src.split("/"),e=t[0],n=t[2],e+"//"+n+"/api/event");function c(t){console.warn("Ignoring Event: "+t)}function r(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return c("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return c("localStorage flag")}catch(t){}var n={};n.n=t,n.u=e&&e.u?e.u:i.href,n.d=o.getAttribute("data-domain"),n.r=a.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props),n.h=1;var r=new XMLHttpRequest;r.open("POST",p,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(n)),r.onreadystatechange=function(){4===r.readyState&&e&&e.callback&&e.callback()}}}var l=window.vince&&window.vince.q||[];window.vince=r;for(var s=0;s<l.length;s++)r.apply(this,l[s]);var u=1;function d(t){if("auxclick"!==t.type||t.button===u){var e,n=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target),r=n&&n.href&&n.href.split("?")[0];return(e=n)&&e.href&&e.host&&e.host!==i.host?f(t,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(t){if(!t)return!1;var e=t.split(".").pop();return g.some(function(t){return t===e})}(r)?f(t,n,{name:"File Download",props:{url:r}}):void 0}}function f(t,e,n){var r=!1;function i(){r||(r=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),r=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&r}}(t,e)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:i}),setTimeout(i,5e3),t.preventDefault())}a.addEventListener("click",d),a.addEventListener("auxclick",d);var v=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],w=o.getAttribute("file-types"),m=o.getAttribute("add-file-types"),g=w&&w.split(",")||m&&m.split(",").concat(v)||v}();