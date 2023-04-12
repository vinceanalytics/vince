!function(){"use strict";var e,t,n,u=window.location,l=window.document,s=l.getElementById("vince"),d=s.getAttribute("data-api")||(e=s.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function f(e){console.warn("Ignoring Event: "+e)}function r(e,t){try{if("true"===window.localStorage.vince_ignore)return f("localStorage flag")}catch(e){}var n=s&&s.getAttribute("data-include"),r=s&&s.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),i=r&&r.split(",").some(o);if(!a||i)return f("exclusion rule")}function o(e){return u.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var p={};p.n=e,p.u=t&&t.u?t.u:u.href,p.d=s.getAttribute("data-domain"),p.r=l.referrer||null,p.w=window.innerWidth,t&&t.meta&&(p.m=JSON.stringify(t.meta)),t&&t.props&&(p.p=t.props);var c=new XMLHttpRequest;c.open("POST",d,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(p)),c.onreadystatechange=function(){4===c.readyState&&t&&t.callback&&t.callback()}}var a=window.vince&&window.vince.q||[];window.vince=r;for(var i=0;i<a.length;i++)r.apply(this,a[i]);var o=1;function p(e){if("auxclick"!==e.type||e.button===o){var t,n=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target),r=n&&n.href&&n.href.split("?")[0];return(t=n)&&t.href&&t.host&&t.host!==u.host?c(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(e){if(!e)return!1;var t=e.split(".").pop();return w.some(function(e){return e===t})}(r)?c(e,n,{name:"File Download",props:{url:r}}):void 0}}function c(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}l.addEventListener("click",p),l.addEventListener("auxclick",p);var v=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],g=s.getAttribute("file-types"),m=s.getAttribute("add-file-types"),w=g&&g.split(",")||m&&m.split(",").concat(v)||v}();