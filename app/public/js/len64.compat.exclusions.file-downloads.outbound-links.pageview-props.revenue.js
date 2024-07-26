!function(){"use strict";var l=window.location,p=window.document,s=p.getElementById("plausible"),u=s.getAttribute("data-api")||(w=(w=s).src.split("/"),o=w[0],w=w[2],o+"//"+w+"/api/event");function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return c("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return c(null,t);try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var a=s&&s.getAttribute("data-include"),n=s&&s.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(i),n=n&&n.split(",").some(i);if(!a||n)return c("exclusion rule",t)}function i(e){return l.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},n=(a.n=e,a.u=l.href,a.d=s.getAttribute("data-domain"),a.r=p.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),r=a.p||{},o=(n.forEach(function(e){var t=e.replace("event-",""),e=s.getAttribute(e);r[t]=r[t]||e}),a.p=r,new XMLHttpRequest);o.open("POST",u,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(a)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function i(){a!==l.pathname&&(a=l.pathname,e("pageview"))}var r,o=window.history;o.pushState&&(r=o.pushState,o.pushState=function(){r.apply(this,arguments),i()},window.addEventListener("popstate",i)),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){a||"visible"!==p.visibilityState||i()}):i();var d=1;function f(e){var t,a,n,i;if("auxclick"!==e.type||e.button===d)return t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],(n=t)&&n.href&&n.host&&n.host!==l.host?v(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):(n=a)&&(i=n.split(".").pop(),h.some(function(e){return e===i}))?v(e,t,{name:"File Download",props:{url:a}}):void 0}function v(e,t,a){var n,i=!1;function r(){i||(i=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((n={props:a.props}).revenue=a.revenue,plausible(a.name,n)):((n={props:a.props,callback:r}).revenue=a.revenue,plausible(a.name,n),setTimeout(r,5e3),e.preventDefault())}p.addEventListener("click",f),p.addEventListener("auxclick",f);var w=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],g=s.getAttribute("file-types"),m=s.getAttribute("add-file-types"),h=g&&g.split(",")||m&&m.split(",").concat(w)||w}();