!function(){"use strict";var i=window.location,o=window.document,l=o.getElementById("plausible"),p=l.getAttribute("data-api")||(d=(d=l).src.split("/"),n=d[0],d=d[2],n+"//"+d+"/api/event");function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return s("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,t);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var r={},e=(r.n=e,r.u=t&&t.u?t.u:i.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),a=r.p||{},n=(e.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);a[t]=a[t]||e}),r.p=a,r.h=1,new XMLHttpRequest);n.open("POST",p,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(r)),n.onreadystatechange=function(){4===n.readyState&&t&&t.callback&&t.callback({status:n.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function u(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var c=1;function a(e){if("auxclick"!==e.type||e.button===c){var t,r,a=function(e){for(;e&&(void 0===e.tagName||!u(e)||!e.href);)e=e.parentNode;return e}(e.target),n=a&&a.href&&a.href.split("?")[0];if(!function e(t,r){if(!t||w<r)return!1;if(h(t))return!0;return e(t.parentNode,r+1)}(a,0))return(t=a)&&t.href&&t.host&&t.host!==i.host?f(e,a,{name:"Outbound Link: Click",props:{url:a.href}}):(t=n)&&(r=t.split(".").pop(),v.some(function(e){return e===r}))?f(e,a,{name:"File Download",props:{url:n}}):void 0}}function f(e,t,r){var a,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(a={props:r.props},plausible(r.name,a)):(a={props:r.props,callback:i},plausible(r.name,a),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",a),o.addEventListener("auxclick",a);var n=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],d=l.getAttribute("file-types"),m=l.getAttribute("add-file-types"),v=d&&d.split(",")||m&&m.split(",").concat(n)||n;function g(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{}},r=e&&e.classList;if(r)for(var a=0;a<r.length;a++){var n,i=r.item(a).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}var w=3;function b(e){if("auxclick"!==e.type||e.button===c){for(var t,r,a,n,i=e.target,o=0;o<=w&&i;o++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;u(i)&&(t=i),h(i)&&(r=i),i=i.parentNode}r&&(n=g(r),t?(n.props.url=t.href,f(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,r=e.target,a=g(r);function n(){t||(t=!0,r.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:a.props,callback:n},plausible(a.name,e))}),o.addEventListener("click",b),o.addEventListener("auxclick",b)}();