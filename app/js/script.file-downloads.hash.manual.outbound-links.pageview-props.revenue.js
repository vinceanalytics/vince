!function(){"use strict";var i=window.location,o=window.document,l=o.currentScript,p=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var r={},e=(r.n=e,r.u=t&&t.u?t.u:i.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),n=r.p||{},a=(e.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);n[t]=n[t]||e}),r.p=n,r.h=1,new XMLHttpRequest);a.open("POST",p,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(r)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback({status:a.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);var s=1;function n(e){var t,r,n,a;if("auxclick"!==e.type||e.button===s)return t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),r=t&&t.href&&t.href.split("?")[0],(n=t)&&n.href&&n.host&&n.host!==i.host?c(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):(n=r)&&(a=n.split(".").pop(),v.some(function(e){return e===a}))?c(e,t,{name:"File Download",props:{url:r}}):void 0}function c(e,t,r){var n,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((n={props:r.props}).revenue=r.revenue,plausible(r.name,n)):((n={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,n),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",n),o.addEventListener("auxclick",n);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],f=l.getAttribute("file-types"),d=l.getAttribute("add-file-types"),v=f&&f.split(",")||d&&d.split(",").concat(a)||a}();