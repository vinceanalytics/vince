!function(){"use strict";var i=window.location,o=window.document,l=o.getElementById("plausible"),p=l.getAttribute("data-api")||(c=(c=l).src.split("/"),n=c[0],c=c[2],n+"//"+c+"/api/event");function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return a=t,(r="localStorage flag")&&console.warn("Ignoring Event: "+r),void(a&&a.callback&&a.callback())}catch(e){}var a,r={},n=(r.n=e,r.u=t&&t.u?t.u:i.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),r.h=1,new XMLHttpRequest);n.open("POST",p,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(r)),n.onreadystatechange=function(){4===n.readyState&&t&&t.callback&&t.callback({status:n.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a=0;a<t.length;a++)e.apply(this,t[a]);var u=1;function r(e){var t,a,r,n;if("auxclick"!==e.type||e.button===u)return t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],(r=t)&&r.href&&r.host&&r.host!==i.host?s(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):(r=a)&&(n=r.split(".").pop(),f.some(function(e){return e===n}))?s(e,t,{name:"File Download",props:{url:a}}):void 0}function s(e,t,a){var r,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((r={props:a.props}).revenue=a.revenue,plausible(a.name,r)):((r={props:a.props,callback:i}).revenue=a.revenue,plausible(a.name,r),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",r),o.addEventListener("auxclick",r);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],c=l.getAttribute("file-types"),d=l.getAttribute("add-file-types"),f=c&&c.split(",")||d&&d.split(",").concat(n)||n}();