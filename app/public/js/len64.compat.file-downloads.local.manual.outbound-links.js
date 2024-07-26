!function(){"use strict";var i=window.location,o=window.document,l=o.getElementById("plausible"),p=l.getAttribute("data-api")||(c=(c=l).src.split("/"),n=c[0],c=c[2],n+"//"+c+"/api/event");function t(t,e){try{if("true"===window.localStorage.plausible_ignore)return a=e,(r="localStorage flag")&&console.warn("Ignoring Event: "+r),void(a&&a.callback&&a.callback())}catch(t){}var a,r={},n=(r.n=t,r.u=e&&e.u?e.u:i.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,e&&e.meta&&(r.m=JSON.stringify(e.meta)),e&&e.props&&(r.p=e.props),new XMLHttpRequest);n.open("POST",p,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(r)),n.onreadystatechange=function(){4===n.readyState&&e&&e.callback&&e.callback({status:n.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a=0;a<e.length;a++)t.apply(this,e[a]);var s=1;function r(t){var e,a,r,n;if("auxclick"!==t.type||t.button===s)return e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(r=e)&&r.href&&r.host&&r.host!==i.host?u(t,e,{name:"Outbound Link: Click",props:{url:e.href}}):(r=a)&&(n=r.split(".").pop(),f.some(function(t){return t===n}))?u(t,e,{name:"File Download",props:{url:a}}):void 0}function u(t,e,a){var r,n=!1;function i(){n||(n=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(t,e)?(r={props:a.props},plausible(a.name,r)):(r={props:a.props,callback:i},plausible(a.name,r),setTimeout(i,5e3),t.preventDefault())}o.addEventListener("click",r),o.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],c=l.getAttribute("file-types"),d=l.getAttribute("add-file-types"),f=c&&c.split(",")||d&&d.split(",").concat(n)||n}();