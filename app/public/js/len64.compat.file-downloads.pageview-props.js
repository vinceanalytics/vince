!function(){"use strict";var r=window.location,o=window.document,p=o.getElementById("plausible"),l=p.getAttribute("data-api")||(w=(w=p).src.split("/"),c=w[0],w=w[2],c+"//"+w+"/api/event");function s(t,e){t&&console.warn("Ignoring Event: "+t),e&&e.callback&&e.callback()}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return s("localhost",e);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,e);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",e)}catch(t){}var a={},t=(a.n=t,a.u=r.href,a.d=p.getAttribute("data-domain"),a.r=o.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),p.getAttributeNames().filter(function(t){return"event-"===t.substring(0,6)})),i=a.p||{},n=(t.forEach(function(t){var e=t.replace("event-",""),t=p.getAttribute(t);i[e]=i[e]||t}),a.p=i,new XMLHttpRequest);n.open("POST",l,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(a)),n.onreadystatechange=function(){4===n.readyState&&e&&e.callback&&e.callback({status:n.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a,i=0;i<e.length;i++)t.apply(this,e[i]);function n(){a!==r.pathname&&(a=r.pathname,t("pageview"))}var u,c=window.history;c.pushState&&(u=c.pushState,c.pushState=function(){u.apply(this,arguments),n()},window.addEventListener("popstate",n)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){a||"visible"!==o.visibilityState||n()}):n();var d=1;function f(t){var e,a,i,n,r,o,p;function l(){n||(n=!0,window.location=i.href)}"auxclick"===t.type&&t.button!==d||(e=function(t){for(;t&&(void 0===t.tagName||!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase()||!t.href);)t=t.parentNode;var e;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],(o=a)&&(p=o.split(".").pop(),m.some(function(t){return t===p}))&&(n=!(o={name:"File Download",props:{url:a}}),!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(a=t,i=e)?(r={props:o.props},plausible(o.name,r)):(r={props:o.props,callback:l},plausible(o.name,r),setTimeout(l,5e3),a.preventDefault())))}o.addEventListener("click",f),o.addEventListener("auxclick",f);var w=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],v=p.getAttribute("file-types"),g=p.getAttribute("add-file-types"),m=v&&v.split(",")||g&&g.split(",").concat(w)||w}();