!function(){"use strict";var l=window.location,p=window.document,o=p.getElementById("plausible"),s=o.getAttribute("data-api")||(d=(d=o).src.split("/"),u=d[0],d=d[2],u+"//"+d+"/api/event");function t(t,e){try{if("true"===window.localStorage.plausible_ignore)return i=e,(a="localStorage flag")&&console.warn("Ignoring Event: "+a),void(i&&i.callback&&i.callback())}catch(t){}var a={},i=(a.n=t,a.u=l.href,a.d=o.getAttribute("data-domain"),a.r=p.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),o.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)})),n=a.p||{},r=(i.forEach(function(t){var e=t.replace("event-",""),t=o.getAttribute(t);n[e]=n[e]||t}),a.p=n,a.h=1,new XMLHttpRequest);r.open("POST",s,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&e&&e.callback&&e.callback({status:r.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a,i=0;i<e.length;i++)t.apply(this,e[i]);function n(){a=l.pathname,t("pageview")}window.addEventListener("hashchange",n),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){a||"visible"!==p.visibilityState||n()}):n();var c=1;function r(t){var e,a,i,n,r,l,p;function o(){n||(n=!0,window.location=i.href)}"auxclick"===t.type&&t.button!==c||(e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(l=a)&&(p=l.split(".").pop(),g.some(function(t){return t===p}))&&(n=!(l={name:"File Download",props:{url:a}}),!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(a=t,i=e)?(r={props:l.props},plausible(l.name,r)):(r={props:l.props,callback:o},plausible(l.name,r),setTimeout(o,5e3),a.preventDefault())))}p.addEventListener("click",r),p.addEventListener("auxclick",r);var u=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],d=o.getAttribute("file-types"),f=o.getAttribute("add-file-types"),g=d&&d.split(",")||f&&f.split(",").concat(u)||u}();