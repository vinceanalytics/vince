!function(){"use strict";var r=window.location,a=window.document,o=a.currentScript,p=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function e(e,t){try{if("true"===window.localStorage.vince_ignore)return void console.warn("Ignoring Event: localStorage flag")}catch(e){}var n={};n.n=e,n.u=r.href,n.d=o.getAttribute("data-domain"),n.r=a.referrer||null,n.w=window.innerWidth,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),n.h=1;var i=new XMLHttpRequest;i.open("POST",p,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,i=0;i<t.length;i++)e.apply(this,t[i]);function c(){n=r.pathname,e("pageview")}window.addEventListener("hashchange",c),"prerender"===a.visibilityState?a.addEventListener("visibilitychange",function(){n||"visible"!==a.visibilityState||c()}):c();var s=1;function l(e){if("auxclick"!==e.type||e.button===s){var t,n=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target),i=n&&n.href&&n.href.split("?")[0];return(t=n)&&t.href&&t.host&&t.host!==r.host?d(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(e){if(!e)return!1;var t=e.split(".").pop();return w.some(function(e){return e===t})}(i)?d(e,n,{name:"File Download",props:{url:i}}):void 0}}function d(e,t,n){var i=!1;function r(){i||(i=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),i=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&i}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:r}),setTimeout(r,5e3),e.preventDefault())}a.addEventListener("click",l),a.addEventListener("auxclick",l);var u=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],v=o.getAttribute("file-types"),f=o.getAttribute("add-file-types"),w=v&&v.split(",")||f&&f.split(",").concat(u)||u}();