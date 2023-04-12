!function(){"use strict";var t,e,n,s=window.location,l=window.document,u=l.getElementById("vince"),d=u.getAttribute("data-api")||(t=u.src.split("/"),e=t[0],n=t[2],e+"//"+n+"/api/event");function f(t){console.warn("Ignoring Event: "+t)}function i(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return f("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return f("localStorage flag")}catch(t){}var n=u&&u.getAttribute("data-include"),i=u&&u.getAttribute("data-exclude");if("pageview"===t){var a=!n||n&&n.split(",").some(c),r=i&&i.split(",").some(c);if(!a||r)return f("exclusion rule")}var o={};o.n=t,o.u=s.href,o.d=u.getAttribute("data-domain"),o.r=l.referrer||null,o.w=window.innerWidth,e&&e.meta&&(o.m=JSON.stringify(e.meta)),e&&e.props&&(o.p=e.props);var p=new XMLHttpRequest;p.open("POST",d,!0),p.setRequestHeader("Content-Type","text/plain"),p.send(JSON.stringify(o)),p.onreadystatechange=function(){4===p.readyState&&e&&e.callback&&e.callback()}}function c(t){return s.pathname.match(new RegExp("^"+t.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var a=window.vince&&window.vince.q||[];window.vince=i;for(var r,o=0;o<a.length;o++)i.apply(this,a[o]);function p(){r!==s.pathname&&(r=s.pathname,i("pageview"))}var c,v=window.history;v.pushState&&(c=v.pushState,v.pushState=function(){c.apply(this,arguments),p()},window.addEventListener("popstate",p)),"prerender"===l.visibilityState?l.addEventListener("visibilitychange",function(){r||"visible"!==l.visibilityState||p()}):p();var w=1;function g(t){if("auxclick"!==t.type||t.button===w){var e,n=function(t){for(;t&&(void 0===t.tagName||(!(e=t)||!e.tagName||"a"!==e.tagName.toLowerCase())||!t.href);)t=t.parentNode;var e;return t}(t.target),i=n&&n.href&&n.href.split("?")[0];return(e=n)&&e.href&&e.host&&e.host!==s.host?m(t,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(t){if(!t)return!1;var e=t.split(".").pop();return x.some(function(t){return t===e})}(i)?m(t,n,{name:"File Download",props:{url:i}}):void 0}}function m(t,e,n){var i=!1;function a(){i||(i=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented){var n=!e.target||e.target.match(/^_(self|parent|top)$/i),i=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type;return n&&i}}(t,e)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),t.preventDefault())}l.addEventListener("click",g),l.addEventListener("auxclick",g);var h=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],y=u.getAttribute("file-types"),b=u.getAttribute("add-file-types"),x=y&&y.split(",")||b&&b.split(",").concat(h)||h}();