!function(){"use strict";var s=window.location,f=window.document,d=f.currentScript,v=d.getAttribute("data-api")||new URL(d.src).origin+"/api/event";function g(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return g("localStorage flag")}catch(e){}var r=d&&d.getAttribute("data-include"),n=d&&d.getAttribute("data-exclude");if("pageview"===e){var a=!r||r&&r.split(",").some(o),i=n&&n.split(",").some(o);if(!a||i)return g("exclusion rule")}function o(e){return s.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=t&&t.u?t.u:s.href,c.d=d.getAttribute("data-domain"),c.r=f.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var p=d.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),u=c.p||{};p.forEach(function(e){var t=e.replace("event-",""),r=d.getAttribute(e);u[t]=u[t]||r}),c.p=u;var l=new XMLHttpRequest;l.open("POST",v,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(c)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);var p=1;function n(e){if("auxclick"!==e.type||e.button===p){var t,r,n,a,i=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target),o=i&&i.href&&i.href.split("?")[0];if(function(e){if(!e)return!1;var t=e.split(".").pop();return u.some(function(e){return e===t})}(o))return a=!(n={name:"File Download",props:{url:o}}),void(!function(e,t){if(!e.defaultPrevented){var r=!t.target||t.target.match(/^_(self|parent|top)$/i),n=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return r&&n}}(t=e,r=i)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=r.href)}}f.addEventListener("click",n),f.addEventListener("auxclick",n);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],i=d.getAttribute("file-types"),o=d.getAttribute("add-file-types"),u=i&&i.split(",")||o&&o.split(",").concat(a)||a}();