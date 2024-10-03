!function(){"use strict";var p=window.location,o=window.document,u=o.getElementById("plausible"),s=u.getAttribute("data-api")||(i=(i=u).src.split("/"),n=i[0],i=i[2],n+"//"+i+"/api/event");function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var a=u&&u.getAttribute("data-include"),r=u&&u.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(n),r=r&&r.split(",").some(n);if(!a||r)return c("exclusion rule",t)}function n(e){var t=p.pathname;return(t+=p.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},r=(a.n=e,a.u=t&&t.u?t.u:p.href,a.d=u.getAttribute("data-domain"),a.r=o.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),u.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=a.p||{},l=(r.forEach(function(e){var t=e.replace("event-",""),e=u.getAttribute(e);i[t]=i[t]||e}),a.p=i,a.h=1,new XMLHttpRequest);l.open("POST",s,!0),l.setRequestHeader("Content-Type","text/plain"),l.send(JSON.stringify(a)),l.onreadystatechange=function(){4===l.readyState&&t&&t.callback&&t.callback({status:l.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a=0;a<t.length;a++)e.apply(this,t[a]);var f=1;function r(e){var t,a,r,n,i,l,p;function o(){n||(n=!0,window.location=r.href)}"auxclick"===e.type&&e.button!==f||(t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],(l=a)&&(p=l.split(".").pop(),d.some(function(e){return e===p}))&&(n=!(l={name:"File Download",props:{url:a}}),!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(a=e,r=t)?((i={props:l.props}).revenue=l.revenue,plausible(l.name,i)):((i={props:l.props,callback:o}).revenue=l.revenue,plausible(l.name,i),setTimeout(o,5e3),a.preventDefault())))}o.addEventListener("click",r),o.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],i=u.getAttribute("file-types"),l=u.getAttribute("add-file-types"),d=i&&i.split(",")||l&&l.split(",").concat(n)||n}();