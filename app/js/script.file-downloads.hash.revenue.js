!function(){"use strict";var i=window.location,r=window.document,o=r.currentScript,l=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return p("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return p(null,t);try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var a={},n=(a.n=e,a.u=i.href,a.d=o.getAttribute("data-domain"),a.r=r.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),a.h=1,new XMLHttpRequest);n.open("POST",l,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(a)),n.onreadystatechange=function(){4===n.readyState&&t&&t.callback&&t.callback({status:n.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function s(){a=i.pathname,e("pageview")}window.addEventListener("hashchange",s),"prerender"===r.visibilityState?r.addEventListener("visibilitychange",function(){a||"visible"!==r.visibilityState||s()}):s();var c=1;function u(e){var t,a,n,i,r,o,l;function p(){i||(i=!0,window.location=n.href)}"auxclick"===e.type&&e.button!==c||(t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],(o=a)&&(l=o.split(".").pop(),v.some(function(e){return e===l}))&&(i=!(o={name:"File Download",props:{url:a}}),!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(a=e,n=t)?((r={props:o.props}).revenue=o.revenue,plausible(o.name,r)):((r={props:o.props,callback:p}).revenue=o.revenue,plausible(o.name,r),setTimeout(p,5e3),a.preventDefault())))}r.addEventListener("click",u),r.addEventListener("auxclick",u);var d=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],w=o.getAttribute("file-types"),f=o.getAttribute("add-file-types"),v=w&&w.split(",")||f&&f.split(",").concat(d)||d}();