!function(){"use strict";var i=window.location,o=window.document,p=o.currentScript,l=p.getAttribute("data-api")||new URL(p.src).origin+"/api/event";function t(t,e){try{if("true"===window.localStorage.plausible_ignore)return a=e,(r="localStorage flag")&&console.warn("Ignoring Event: "+r),void(a&&a.callback&&a.callback())}catch(t){}var a,r={},n=(r.n=t,r.u=e&&e.u?e.u:i.href,r.d=p.getAttribute("data-domain"),r.r=o.referrer||null,e&&e.meta&&(r.m=JSON.stringify(e.meta)),e&&e.props&&(r.p=e.props),r.h=1,new XMLHttpRequest);n.open("POST",l,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(r)),n.onreadystatechange=function(){4===n.readyState&&e&&e.callback&&e.callback({status:n.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a=0;a<e.length;a++)t.apply(this,e[a]);var c=1;function r(t){var e,a,r,n,i,o,p;function l(){n||(n=!0,window.location=r.href)}"auxclick"===t.type&&t.button!==c||(e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(o=a)&&(p=o.split(".").pop(),d.some(function(t){return t===p}))&&(n=!(o={name:"File Download",props:{url:a}}),!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(a=t,r=e)?(i={props:o.props},plausible(o.name,i)):(i={props:o.props,callback:l},plausible(o.name,i),setTimeout(l,5e3),a.preventDefault())))}o.addEventListener("click",r),o.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],s=p.getAttribute("file-types"),u=p.getAttribute("add-file-types"),d=s&&s.split(",")||u&&u.split(",").concat(n)||n}();