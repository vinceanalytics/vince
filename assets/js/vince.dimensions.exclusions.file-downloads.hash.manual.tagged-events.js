!function(){"use strict";var l=window.location,f=window.document,v=f.currentScript,d=v.getAttribute("data-api")||new URL(v.src).origin+"/api/event";function m(e){console.warn("Ignoring Event: "+e)}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(l.hostname)||"file:"===l.protocol)return m("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return m("localStorage flag")}catch(e){}var r=v&&v.getAttribute("data-include"),n=v&&v.getAttribute("data-exclude");if("pageview"===e){var a=!r||r&&r.split(",").some(s),i=n&&n.split(",").some(s);if(!a||i)return m("exclusion rule")}var o={};o.n=e,o.u=t&&t.u?t.u:l.href,o.d=v.getAttribute("data-domain"),o.r=f.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props);var c=v.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),p=o.p||{};c.forEach(function(e){var t=e.replace("event-",""),r=v.getAttribute(e);p[t]=p[t]||r}),o.p=p,o.h=1;var u=new XMLHttpRequest;u.open("POST",d,!0),u.setRequestHeader("Content-Type","text/plain"),u.send(JSON.stringify(o)),u.onreadystatechange=function(){4===u.readyState&&t&&t.callback&&t.callback()}}function s(e){var t=l.pathname;return(t+=l.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var p=1;function n(e){if("auxclick"!==e.type||e.button===p){var t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),r=t&&t.href&&t.href.split("?")[0];if(!function e(t,r){if(!t||w<r)return!1;if(b(t))return!0;return e(t.parentNode,r+1)}(t,0))return function(e){if(!e)return!1;var t=e.split(".").pop();return s.some(function(e){return e===t})}(r)?u(e,t,{name:"File Download",props:{url:r}}):void 0}}function u(e,t,r){var n=!1;function a(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var r=!t.target||t.target.match(/^_(self|parent|top)$/i),n=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return r&&n}}(e,t)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}f.addEventListener("click",n),f.addEventListener("auxclick",n);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],i=v.getAttribute("file-types"),o=v.getAttribute("add-file-types"),s=i&&i.split(",")||o&&o.split(",").concat(a)||a;function g(e){var t=b(e)?e:e&&e.parentNode,r={name:null,props:{}},n=t&&t.classList;if(!n)return r;for(var a=0;a<n.length;a++){var i,o,c=n.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?r.name=o:r.props[i]=o)}return r}var w=3;function h(e){if("auxclick"!==e.type||e.button===p){for(var t,r,n,a,i=e.target,o=0;o<=w&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;c(i)&&(t=i),b(i)&&(r=i),i=i.parentNode}r&&(a=g(r),t?(a.props.url=t.href,u(e,t,a)):vince(a.name,{props:a.props}))}}function b(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/vince-event-name=(.+)/))return 1}f.addEventListener("submit",function(e){var t,r=e.target,n=g(r);function a(){t||(t=!0,r.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(n.name,{props:n.props,callback:a}))}),f.addEventListener("click",h),f.addEventListener("auxclick",h)}();