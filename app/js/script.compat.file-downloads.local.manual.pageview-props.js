!function(){"use strict";var p=window.location,l=window.document,o=l.getElementById("plausible"),s=o.getAttribute("data-api")||(i=(i=o).src.split("/"),n=i[0],i=i[2],n+"//"+i+"/api/event");function t(t,e){try{if("true"===window.localStorage.plausible_ignore)return r=e,(a="localStorage flag")&&console.warn("Ignoring Event: "+a),void(r&&r.callback&&r.callback())}catch(t){}var a={},r=(a.n=t,a.u=e&&e.u?e.u:p.href,a.d=o.getAttribute("data-domain"),a.r=l.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),o.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)})),n=a.p||{},i=(r.forEach(function(t){var e=t.replace("event-",""),t=o.getAttribute(t);n[e]=n[e]||t}),a.p=n,new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&e&&e.callback&&e.callback({status:i.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a=0;a<e.length;a++)t.apply(this,e[a]);var c=1;function r(t){var e,a,r,n,i,p,l;function o(){n||(n=!0,window.location=r.href)}"auxclick"===t.type&&t.button!==c||(e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(p=a)&&(l=p.split(".").pop(),d.some(function(t){return t===l}))&&(n=!(p={name:"File Download",props:{url:a}}),!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(a=t,r=e)?(i={props:p.props},plausible(p.name,i)):(i={props:p.props,callback:o},plausible(p.name,i),setTimeout(o,5e3),a.preventDefault())))}l.addEventListener("click",r),l.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=o.getAttribute("file-types"),u=o.getAttribute("add-file-types"),d=i&&i.split(",")||u&&u.split(",").concat(n)||n}();