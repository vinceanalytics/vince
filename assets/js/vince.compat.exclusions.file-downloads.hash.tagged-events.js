!function(){"use strict";var e,t,n,s=window.location,u=window.document,l=u.getElementById("vince"),f=l.getAttribute("data-api")||(e=l.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function v(e){console.warn("Ignoring Event: "+e)}function r(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(s.hostname)||"file:"===s.protocol)return v("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){var i=!n||n&&n.split(",").some(p),a=r&&r.split(",").some(p);if(!i||a)return v("exclusion rule")}var o={};o.n=e,o.u=s.href,o.d=l.getAttribute("data-domain"),o.r=u.referrer||null,o.w=window.innerWidth,t&&t.meta&&(o.m=JSON.stringify(t.meta)),t&&t.props&&(o.p=t.props),o.h=1;var c=new XMLHttpRequest;c.open("POST",f,!0),c.setRequestHeader("Content-Type","text/plain"),c.send(JSON.stringify(o)),c.onreadystatechange=function(){4===c.readyState&&t&&t.callback&&t.callback()}}function p(e){var t=s.pathname;return(t+=s.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}}var i=window.vince&&window.vince.q||[];window.vince=r;for(var a,o=0;o<i.length;o++)r.apply(this,i[o]);function c(){a=s.pathname,r("pageview")}function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",c),"prerender"===u.visibilityState?u.addEventListener("visibilitychange",function(){a||"visible"!==u.visibilityState||c()}):c();var d=1;function m(e){if("auxclick"!==e.type||e.button===d){var t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target),n=t&&t.href&&t.href.split("?")[0];if(!function e(t,n){if(!t||k<n)return!1;if(N(t))return!0;return e(t.parentNode,n+1)}(t,0))return function(e){if(!e)return!1;var t=e.split(".").pop();return b.some(function(e){return e===t})}(n)?g(e,t,{name:"File Download",props:{url:n}}):void 0}}function g(e,t,n){var r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:i}),setTimeout(i,5e3),e.preventDefault())}u.addEventListener("click",m),u.addEventListener("auxclick",m);var w=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],h=l.getAttribute("file-types"),y=l.getAttribute("add-file-types"),b=h&&h.split(",")||y&&y.split(",").concat(w)||w;function x(e){var t=N(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var i=0;i<r.length;i++){var a,o,c=r.item(i).match(/vince-event-(.+)=(.+)/);c&&(a=c[1],o=c[2].replace(/\+/g," "),"name"===a.toLowerCase()?n.name=o:n.props[a]=o)}return n}var k=3;function L(e){if("auxclick"!==e.type||e.button===d){for(var t,n,r,i,a=e.target,o=0;o<=k&&a;o++){if((r=a)&&r.tagName&&"form"===r.tagName.toLowerCase())return;p(a)&&(t=a),N(a)&&(n=a),a=a.parentNode}n&&(i=x(n),t?(i.props.url=t.href,g(e,t,i)):vince(i.name,{props:i.props}))}}function N(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}u.addEventListener("submit",function(e){var t,n=e.target,r=x(n);function i(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(i,5e3),vince(r.name,{props:r.props,callback:i}))}),u.addEventListener("click",L),u.addEventListener("auxclick",L)}();