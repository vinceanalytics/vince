!function(){"use strict";var e,t,n,o=window.location,c=window.document,p=c.getElementById("vince"),s=p.getAttribute("data-api")||(e=p.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function u(e){console.warn("Ignoring Event: "+e)}function r(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return u("localhost");if(!(window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)){try{if("true"===window.localStorage.vince_ignore)return u("localStorage flag")}catch(e){}var n={};n.n=e,n.u=o.href,n.d=p.getAttribute("data-domain"),n.r=c.referrer||null,n.w=window.innerWidth,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props);var r=p.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),i=n.p||{};r.forEach(function(e){var t=e.replace("event-",""),n=p.getAttribute(e);i[t]=i[t]||n}),n.p=i,n.h=1;var a=new XMLHttpRequest;a.open("POST",s,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback()}}}var i=window.vince&&window.vince.q||[];window.vince=r;for(var a,l=0;l<i.length;l++)r.apply(this,i[l]);function f(){a=o.pathname,r("pageview")}function v(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",f),"prerender"===c.visibilityState?c.addEventListener("visibilitychange",function(){a||"visible"!==c.visibilityState||f()}):f();var d=1;function m(e){if("auxclick"!==e.type||e.button===d){var t,n=function(e){for(;e&&(void 0===e.tagName||!v(e)||!e.href);)e=e.parentNode;return e}(e.target),r=n&&n.href&&n.href.split("?")[0];if(!function e(t,n){if(!t||L<n)return!1;if(x(t))return!0;return e(t.parentNode,n+1)}(n,0))return(t=n)&&t.href&&t.host&&t.host!==o.host?g(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(e){if(!e)return!1;var t=e.split(".").pop();return y.some(function(e){return e===t})}(r)?g(e,n,{name:"File Download",props:{url:r}}):void 0}}function g(e,t,n){var r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:i}),setTimeout(i,5e3),e.preventDefault())}c.addEventListener("click",m),c.addEventListener("auxclick",m);var w=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],h=p.getAttribute("file-types"),b=p.getAttribute("add-file-types"),y=h&&h.split(",")||b&&b.split(",").concat(w)||w;function k(e){var t=x(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var i=0;i<r.length;i++){var a,o,c=r.item(i).match(/vince-event-(.+)=(.+)/);c&&(a=c[1],o=c[2].replace(/\+/g," "),"name"===a.toLowerCase()?n.name=o:n.props[a]=o)}return n}var L=3;function N(e){if("auxclick"!==e.type||e.button===d){for(var t,n,r,i,a=e.target,o=0;o<=L&&a;o++){if((r=a)&&r.tagName&&"form"===r.tagName.toLowerCase())return;v(a)&&(t=a),x(a)&&(n=a),a=a.parentNode}n&&(i=k(n),t?(i.props.url=t.href,g(e,t,i)):vince(i.name,{props:i.props}))}}function x(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}c.addEventListener("submit",function(e){var t,n=e.target,r=k(n);function i(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(i,5e3),vince(r.name,{props:r.props,callback:i}))}),c.addEventListener("click",N),c.addEventListener("auxclick",N)}();