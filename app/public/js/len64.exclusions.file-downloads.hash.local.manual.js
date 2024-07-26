!function(){"use strict";var p=window.location,l=window.document,o=l.currentScript,c=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function s(t,e){t&&console.warn("Ignoring Event: "+t),e&&e.callback&&e.callback()}function t(t,e){try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",e)}catch(t){}var a=o&&o.getAttribute("data-include"),r=o&&o.getAttribute("data-exclude");if("pageview"===t){a=!a||a.split(",").some(n),r=r&&r.split(",").some(n);if(!a||r)return s("exclusion rule",e)}function n(t){var e=p.pathname;return(e+=p.hash).match(new RegExp("^"+t.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},i=(a.n=t,a.u=e&&e.u?e.u:p.href,a.d=o.getAttribute("data-domain"),a.r=l.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),a.h=1,new XMLHttpRequest);i.open("POST",c,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&e&&e.callback&&e.callback({status:i.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a=0;a<e.length;a++)t.apply(this,e[a]);var u=1;function r(t){var e,a,r,n,i,p,l;function o(){n||(n=!0,window.location=r.href)}"auxclick"===t.type&&t.button!==u||(e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(p=a)&&(l=p.split(".").pop(),f.some(function(t){return t===l}))&&(n=!(p={name:"File Download",props:{url:a}}),!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(a=t,r=e)?(i={props:p.props},plausible(p.name,i)):(i={props:p.props,callback:o},plausible(p.name,i),setTimeout(o,5e3),a.preventDefault())))}l.addEventListener("click",r),l.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=o.getAttribute("file-types"),d=o.getAttribute("add-file-types"),f=i&&i.split(",")||d&&d.split(",").concat(n)||n}();