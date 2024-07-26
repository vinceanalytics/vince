!function(){"use strict";var r=window.location,o=window.document,l=o.getElementById("plausible"),p=l.getAttribute("data-api")||(v=(v=l).src.split("/"),f=v[0],v=v[2],f+"//"+v+"/api/event");function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return s("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,t);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var n={},e=(n.n=e,n.u=r.href,n.d=l.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),a=n.p||{},i=(e.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);a[t]=a[t]||e}),n.p=a,n.h=1,new XMLHttpRequest);i.open("POST",p,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n,a=0;a<t.length;a++)e.apply(this,t[a]);function i(){n=r.pathname,e("pageview")}window.addEventListener("hashchange",i),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){n||"visible"!==o.visibilityState||i()}):i();var u=1;function c(e){var t,n,a,i;if("auxclick"!==e.type||e.button===u)return t=function(e){for(;e&&(void 0===e.tagName||!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase()||!e.href);)e=e.parentNode;var t;return e}(e.target),n=t&&t.href&&t.href.split("?")[0],(a=t)&&a.href&&a.host&&a.host!==r.host?d(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):(a=n)&&(i=a.split(".").pop(),g.some(function(e){return e===i}))?d(e,t,{name:"File Download",props:{url:n}}):void 0}function d(e,t,n){var a,i=!1;function r(){i||(i=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:n.props}).revenue=n.revenue,plausible(n.name,a)):((a={props:n.props,callback:r}).revenue=n.revenue,plausible(n.name,a),setTimeout(r,5e3),e.preventDefault())}o.addEventListener("click",c),o.addEventListener("auxclick",c);var f=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],v=l.getAttribute("file-types"),w=l.getAttribute("add-file-types"),g=v&&v.split(",")||w&&w.split(",").concat(f)||f}();