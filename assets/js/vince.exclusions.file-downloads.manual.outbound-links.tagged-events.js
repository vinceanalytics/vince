!function(){"use strict";var u=window.location,s=window.document,l=s.currentScript,f=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function v(e){console.warn("Ignoring Event: "+e)}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(u.hostname)||"file:"===u.protocol)return v("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(p),i=r&&r.split(",").some(p);if(!a||i)return v("exclusion rule")}var o={};o.n=e,o.u=t&&t.u?t.u:u.href,o.d=l.getAttribute("data-domain"),o.r=s.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props);var c=new XMLHttpRequest;c.open("POST",f,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(o)),c.onreadystatechange=function(){4===c.readyState&&t&&t.callback&&t.callback()}}function p(e){return u.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n=0;n<t.length;n++)e.apply(this,t[n]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var p=1;function r(e){if("auxclick"!==e.type||e.button===p){var t,n=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),r=n&&n.href&&n.href.split("?")[0];if(!function e(t,n){if(!t||w<n)return!1;if(y(t))return!0;return e(t.parentNode,n+1)}(n,0))return(t=n)&&t.href&&t.host&&t.host!==u.host?d(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(e){if(!e)return!1;var t=e.split(".").pop();return m.some(function(e){return e===t})}(r)?d(e,n,{name:"File Download",props:{url:r}}):void 0}}function d(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}s.addEventListener("click",r),s.addEventListener("auxclick",r);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],i=l.getAttribute("file-types"),o=l.getAttribute("add-file-types"),m=i&&i.split(",")||o&&o.split(",").concat(a)||a;function g(e){var t=y(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var a=0;a<r.length;a++){var i,o,c=r.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}var w=3;function h(e){if("auxclick"!==e.type||e.button===p){for(var t,n,r,a,i=e.target,o=0;o<=w&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),y(i)&&(n=i),i=i.parentNode}n&&(a=g(n),t?(a.props.url=t.href,d(e,t,a)):vince(a.name,{props:a.props}))}}function y(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}s.addEventListener("submit",function(e){var t,n=e.target,r=g(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(r.name,{props:r.props,callback:a}))}),s.addEventListener("click",h),s.addEventListener("auxclick",h)}();