!function(){"use strict";var t,e,n,o=window.location,p=window.document,c=p.getElementById("vince"),l=c.getAttribute("data-api")||(t=c.src.split("/"),e=t[0],n=t[2],e+"//"+n+"/api/event");function s(t){console.warn("Ignoring Event: "+t)}function r(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return s("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return s("localStorage flag")}catch(t){}var n={};n.n=t,n.u=e&&e.u?e.u:o.href,n.d=c.getAttribute("data-domain"),n.r=p.referrer||null,n.w=window.innerWidth,e&&e.meta&&(n.m=JSON.stringify(e.meta)),e&&e.props&&(n.p=e.props);var r=c.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)}),i=n.p||{};r.forEach(function(t){var e=t.replace("event-",""),n=c.getAttribute(t);i[e]=i[e]||n}),n.p=i,n.h=1;var a=new XMLHttpRequest;a.open("POST",l,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&e&&e.callback&&e.callback()}}}var i=window.vince&&window.vince.q||[];window.vince=r;for(var a=0;a<i.length;a++)r.apply(this,i[a]);var u=1;function f(t){if("auxclick"!==t.type||t.button===u){var e,n,r,i,a=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target),o=a&&a.href&&a.href.split("?")[0];if(function(t){if(!t)return!1;var e=t.split(".").pop();return g.some(function(t){return t===e})}(o))return i=!(r={name:"File Download",props:{url:o}}),void(!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),r=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&r}}(e=t,n=a)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:p}),setTimeout(p,5e3),e.preventDefault()))}function p(){i||(i=!0,window.location=n.href)}}p.addEventListener("click",f),p.addEventListener("auxclick",f);var d=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],v=c.getAttribute("file-types"),w=c.getAttribute("add-file-types"),g=v&&v.split(",")||w&&w.split(",").concat(d)||d}();