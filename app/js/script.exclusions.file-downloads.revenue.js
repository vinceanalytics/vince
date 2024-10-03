!function(){"use strict";var o=window.location,p=window.document,l=p.currentScript,s=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var a=l&&l.getAttribute("data-include"),i=l&&l.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(n),i=i&&i.split(",").some(n);if(!a||i)return u("exclusion rule",t)}function n(e){return o.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},r=(a.n=e,a.u=o.href,a.d=l.getAttribute("data-domain"),a.r=p.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),new XMLHttpRequest);r.open("POST",s,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,i=0;i<t.length;i++)e.apply(this,t[i]);function n(){a!==o.pathname&&(a=o.pathname,e("pageview"))}var r,c=window.history;c.pushState&&(r=c.pushState,c.pushState=function(){r.apply(this,arguments),n()},window.addEventListener("popstate",n)),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){a||"visible"!==p.visibilityState||n()}):n();var d=1;function w(e){var t,a,i,n,r,o,p;function l(){n||(n=!0,window.location=i.href)}"auxclick"===e.type&&e.button!==d||(t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],(o=a)&&(p=o.split(".").pop(),g.some(function(e){return e===p}))&&(n=!(o={name:"File Download",props:{url:a}}),!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(a=e,i=t)?((r={props:o.props}).revenue=o.revenue,plausible(o.name,r)):((r={props:o.props,callback:l}).revenue=o.revenue,plausible(o.name,r),setTimeout(l,5e3),a.preventDefault())))}p.addEventListener("click",w),p.addEventListener("auxclick",w);var c=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],f=l.getAttribute("file-types"),v=l.getAttribute("add-file-types"),g=f&&f.split(",")||v&&v.split(",").concat(c)||c}();