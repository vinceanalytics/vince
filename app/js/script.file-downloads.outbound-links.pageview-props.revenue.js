!function(){"use strict";var r=window.location,o=window.document,p=o.currentScript,l=p.getAttribute("data-api")||new URL(p.src).origin+"/api/event";function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return s("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,t);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var n={},e=(n.n=e,n.u=r.href,n.d=p.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),p.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),a=n.p||{},i=(e.forEach(function(e){var t=e.replace("event-",""),e=p.getAttribute(e);a[t]=a[t]||e}),n.p=a,new XMLHttpRequest);i.open("POST",l,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n,a=0;a<t.length;a++)e.apply(this,t[a]);function i(){n!==r.pathname&&(n=r.pathname,e("pageview"))}var u,c=window.history;c.pushState&&(u=c.pushState,c.pushState=function(){u.apply(this,arguments),i()},window.addEventListener("popstate",i)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){n||"visible"!==o.visibilityState||i()}):i();var d=1;function f(e){var t,n,a,i;if("auxclick"!==e.type||e.button===d)return t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),n=t&&t.href&&t.href.split("?")[0],(a=t)&&a.href&&a.host&&a.host!==r.host?v(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):(a=n)&&(i=a.split(".").pop(),h.some(function(e){return e===i}))?v(e,t,{name:"File Download",props:{url:n}}):void 0}function v(e,t,n){var a,i=!1;function r(){i||(i=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:n.props}).revenue=n.revenue,plausible(n.name,a)):((a={props:n.props,callback:r}).revenue=n.revenue,plausible(n.name,a),setTimeout(r,5e3),e.preventDefault())}o.addEventListener("click",f),o.addEventListener("auxclick",f);var c=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],w=p.getAttribute("file-types"),g=p.getAttribute("add-file-types"),h=w&&w.split(",")||g&&g.split(",").concat(c)||c}();