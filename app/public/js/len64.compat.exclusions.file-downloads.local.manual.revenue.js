!function(){"use strict";var l=window.location,p=window.document,o=p.getElementById("plausible"),u=o.getAttribute("data-api")||(i=(i=o).src.split("/"),n=i[0],i=i[2],n+"//"+i+"/api/event");function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var a=o&&o.getAttribute("data-include"),r=o&&o.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(n),r=r&&r.split(",").some(n);if(!a||r)return s("exclusion rule",t)}function n(e){return l.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},i=(a.n=e,a.u=t&&t.u?t.u:l.href,a.d=o.getAttribute("data-domain"),a.r=p.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),new XMLHttpRequest);i.open("POST",u,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a=0;a<t.length;a++)e.apply(this,t[a]);var c=1;function r(e){var t,a,r,n,i,l,p;function o(){n||(n=!0,window.location=r.href)}"auxclick"===e.type&&e.button!==c||(t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],(l=a)&&(p=l.split(".").pop(),f.some(function(e){return e===p}))&&(n=!(l={name:"File Download",props:{url:a}}),!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(a=e,r=t)?((i={props:l.props}).revenue=l.revenue,plausible(l.name,i)):((i={props:l.props,callback:o}).revenue=l.revenue,plausible(l.name,i),setTimeout(o,5e3),a.preventDefault())))}p.addEventListener("click",r),p.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=o.getAttribute("file-types"),d=o.getAttribute("add-file-types"),f=i&&i.split(",")||d&&d.split(",").concat(n)||n}();